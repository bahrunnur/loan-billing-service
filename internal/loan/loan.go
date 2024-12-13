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
		LoanID:                  loanID,
		IsDelinquent:            false,
		LastPaymentDate:         now,
		NextExpectedPaymentDate: now.AddDate(0, 0, 7),
		LateFee:                 currency.NewRupiah(0, 0),
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
	// =====

	return loan, nil
}

// CheckDelinquency check delinquency for a loan based on provided time (or IsDelinquent)
func (ls *LoanService) CheckDelinquency(loanID model.LoanID, when time.Time) (bool, error) {
	// validate time, you can't check with time after the server current time
	// this is needed because there will be a mutable operation that flag the loan account to be delinquent or not
	now := time.Now().UTC()
	when = when.UTC() // making sure

	if when.After(now) {
		return false, model.ErrCheckFutureDelinquent
	}

	loan, err := ls.storage.GetLoanWithDelinquency(loanID)
	if err != nil {
		return false, err
	}

	if ls.isDelinquent(loan.LastPaymentDate, when) {
		err := ls.storage.UpdateLoanDelinquency(loanID, true) // TODO: use better method
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// RecordPayment records a loan payment (or MakePayment)
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
	if loan.IsDelinquent || ls.isDelinquent(loan.LastPaymentDate, when) {
		err := ls.storage.UpdateLoanDelinquency(loanID, true) // TODO: use better method
		if err != nil {
			return err
		}
		return model.ErrPayInDelinquent
	}

	amountNeeded := loan.WeeklyPayment
	missedPayments := 0

	// check if need repayment
	if when.After(loan.NextExpectedPaymentDate) {
		currentCheckDate := loan.LastPaymentDate.AddDate(0, 0, 7)
		for currentCheckDate.Before(when) {
			missedPayments++
			currentCheckDate = currentCheckDate.AddDate(0, 0, 7)
		}

		amountNeeded = amountNeeded.Add(amountNeeded.Multiply(missedPayments))
	}

	// payment has to be exact with the weekly payment multiplier
	if amountNeeded != paymentAmount {
		return model.ErrMismatchPayment
	}

	loanUpdateParams := loan.WeeklyLoan
	delinquencyUpdateParams := loan.DelinquencyStatus

	delinquencyUpdateParams.LastPaymentDate = when

	payment := model.Payment{
		Date:          when,
		Amount:        paymentAmount,
		BalanceBefore: loan.OutstandingBalance,
	}

	if paymentAmount >= loan.OutstandingBalance {
		payment.BalanceAfter = currency.NewRupiah(0, 0)
		payment.LoanCompleted = true
		loanUpdateParams.IsCompleted = true
		delinquencyUpdateParams.NextExpectedPaymentDate = when
	} else {
		payment.BalanceAfter = loan.OutstandingBalance
		loanUpdateParams.OutstandingBalance = loan.OutstandingBalance.Subtract(paymentAmount)
		// calculate next expected payment
		next := loan.NextExpectedPaymentDate.AddDate(0, 0, 7*(missedPayments+1))
		delinquencyUpdateParams.NextExpectedPaymentDate = next
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

	return nil
}
