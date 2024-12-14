package loan

import (
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/model"
	"github.com/bahrunnur/loan-billing-service/internal/ports"
	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	"go.jetify.com/typeid"
)

// Domain or Business Logic for Loan Billing

// LoanStorageAdapter is a consumer interface to interact with storage adapter (repo)
type LoanStorageAdapter interface {
	ports.LoanCreator
	ports.LoanGetter
	ports.LoanUpdater
	ports.DelinquencyStatusCreator
	ports.DelinquencyStatusGetter
	ports.DelinquencyStatusUpdater
	ports.PaymentInserter
	ports.BillingInserter
	ports.BillingGetter
	ports.BillingUpdater
}

// LoanService manages loan-related operations
type LoanService struct {
	storage LoanStorageAdapter
}

// NewLoanService creates a new LoanService
func NewLoanService(storageAdapter LoanStorageAdapter) *LoanService {
	return &LoanService{
		storage: storageAdapter,
	}
}

// GetLoan to get all of the information from that loan including the delinquency status
func (ls *LoanService) GetLoan(loanID model.LoanID) (model.WeeklyLoanFullInformation, error) {
	return ls.storage.GetLoanFullInformation(loanID)
}

// CreateLoan initializes a new loan with weekly payments
func (ls *LoanService) CreateLoan(principal currency.Rupiah, annualInterestRate model.BPS, weeklyLoanTerm int) (model.WeeklyLoan, error) {
	// NOTE: flat (not compound) interest rate: 1000bps (10%)

	// validation, tiger style
	if !(annualInterestRate >= 0) {
		return model.WeeklyLoan{}, model.ErrNegativeInterest
	}

	if !(principal.Rupiah() > 0 || principal.Sen() > 0) {
		return model.WeeklyLoan{}, model.ErrNoPrincipal
	}

	if !(weeklyLoanTerm > 0) {
		return model.WeeklyLoan{}, model.ErrNoTerm
	}

	weeklyPrincipal := principal.Divide(weeklyLoanTerm)
	// TODO: use more precise model like `Decimal`
	weeklyInterest := weeklyPrincipal.Multiply(annualInterestRate.ToPercentage()).Divide(model.PERCENT)
	weeklyPayment := weeklyPrincipal.Add(weeklyInterest)
	totalInterest := weeklyInterest.Multiply(weeklyLoanTerm)
	outstandingBalance := principal.Add(totalInterest)

	loanID, err := typeid.New[model.LoanID]()
	if err != nil {
		return model.WeeklyLoan{}, err
	}

	now := time.Now().UTC()
	loan := model.WeeklyLoan{
		Loan: model.Loan{
			ID:                 loanID,
			Principal:          principal,
			AnnualInterestRate: annualInterestRate,
			StartDate:          now,
			TotalInterest:      totalInterest,
			OutstandingBalance: outstandingBalance,
		},
		LoanTermWeeks:  weeklyLoanTerm,
		WeeklyPayment:  weeklyPayment,
		WeeklyInterest: weeklyInterest,
	}
	delinquencyStatus := model.DelinquencyStatus{
		LoanID:       loanID,
		IsDelinquent: false,
		LateFee:      currency.NewRupiah(0, 0),
	}
	billingParam := model.BillingParam{
		LoanCreationDate: now,
		NumberOfTerm:     weeklyLoanTerm,
		RepaymentAmount:  weeklyPayment,
	}

	// =====
	// TODO: wrap this in sql transaction block
	err = ls.storage.CreateLoan(loan)
	if err != nil {
		return model.WeeklyLoan{}, err
	}

	err = ls.storage.CreateDelinquencyStatus(loanID, delinquencyStatus)
	if err != nil {
		return model.WeeklyLoan{}, err
	}

	err = ls.storage.CreateBilling(loanID, billingParam)
	if err != nil {
		return model.WeeklyLoan{}, err
	}
	// =====

	return loan, nil
}

// CheckDelinquency check delinquency for a loan based on provided time (or IsDelinquent)
func (ls *LoanService) CheckDelinquency(loanID model.LoanID, when time.Time) (bool, error) {
	when = when.UTC() // making sure

	loan, err := ls.storage.GetLoanWithDelinquency(loanID)
	if err != nil {
		return false, err
	}

	// short circuit
	if loan.IsDelinquent {
		return true, nil
	}

	isDelinquent, _, err := ls.ColdDelinquentFlag(loanID, when)
	if err != nil {
		return false, err
	}

	return isDelinquent, nil
}

// RecordPayment records a loan payment (or MakePayment) [idempotent operation]
func (ls *LoanService) RecordPayment(loanID model.LoanID, when time.Time, paymentAmount currency.Rupiah) error {
	when = when.UTC() // make sure, as this service data is in UTC

	loan, err := ls.storage.GetLoanWithDelinquency(loanID)
	if err != nil {
		return err
	}

	// validation
	if loan.IsCompleted {
		return model.ErrRepaymentComplete
	}

	// if delinquent, payment cannot be made, not sure about this as I don't know how delinquent account being handled
	if loan.IsDelinquent {
		return model.ErrPayInDelinquent
	}

	// due dillligence check
	isDelinquent, unfulfilledBilling, err := ls.ColdDelinquentFlag(loanID, when)
	if err != nil {
		return err
	}

	if isDelinquent {
		return model.ErrPayInDelinquent
	}

	amountNeeded := currency.NewRupiah(0, 0)
	missedPayments := 0

	for i, billing := range unfulfilledBilling {
		if i+1 > model.MISSED_PAYMENT_THRESHOLD {
			missedPayments++
		}
		amountNeeded = amountNeeded.Add(billing.Repayment)
	}

	// payment has to be exact with the weekly payment multiplier
	if amountNeeded != paymentAmount {
		return model.ErrMismatchPayment
	}

	loanUpdateParams := loan.WeeklyLoan
	delinquencyUpdateParams := loan.DelinquencyStatus

	payment := model.Payment{
		Date:          when,
		Amount:        paymentAmount,
		BalanceBefore: loan.OutstandingBalance,
	}

	if paymentAmount >= loan.OutstandingBalance {
		payment.BalanceAfter = currency.NewRupiah(0, 0)
		loanUpdateParams.IsCompleted = true
	} else {
		payment.BalanceAfter = loan.OutstandingBalance
		loanUpdateParams.OutstandingBalance = loan.OutstandingBalance.Subtract(paymentAmount)
	}

	err = ls.storage.RecordPayment(loanID, payment)
	if err != nil {
		return err
	}

	err = ls.storage.UpdateLoan(loanID, loanUpdateParams)
	if err != nil {
		return err
	}

	err = ls.storage.UpdateDelinquencyStatus(loanID, delinquencyUpdateParams)
	if err != nil {
		return err
	}

	// update billing status until
	err = ls.storage.PayBillingUntil(loanID, when)
	if err != nil {
		return err
	}

	return nil
}
