package loan

import (
	"fmt"
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/model"
	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	"go.jetify.com/typeid"
)

// Domain or Business Logic for Loan Billing

// LoanService manages loan-related operations
type LoanService struct {
	Loans map[string]*model.WeeklyLoan
}

// NewLoanService creates a new LoanService
func NewLoanService() *LoanService {
	return &LoanService{
		Loans: make(map[string]*model.WeeklyLoan),
	}
}

// CreateLoan initializes a new loan with weekly payments
func (ls *LoanService) CreateLoan(principal currency.Rupiah, annualInterestRate model.BPS, loanTermWeekly int) error {
	// NOTE: flat (not compound) interest rate: 1000bps (10%)

	// validation, tiger style
	if !(annualInterestRate >= 0) {
		return model.ErrNegativeInterest
	}

	if !(principal.Rupiah() > 0 || principal.Sen() > 0) {
		return model.ErrNoPrincipal
	}

	if !(loanTermWeekly > 0) {
		return model.ErrNoTerm
	}

	weeklyPrincipal := principal.Divide(loanTermWeekly)
	// TODO: use more precise model like `Decimal`
	weeklyInterest := weeklyPrincipal.Multiply(annualInterestRate.ToPercentage()).Divide(model.PERCENT)
	weeklyPayment := weeklyPrincipal.Add(weeklyInterest)
	totalInterest := weeklyInterest.Multiply(loanTermWeekly)
	outstandingBalance := principal.Add(totalInterest)

	tid, err := typeid.New[model.LoanID]()
	if err != nil {
		return err
	}

	loan := &model.WeeklyLoan{
		Loan: model.Loan{
			ID:                 tid,
			Principal:          principal,
			AnnualInterestRate: annualInterestRate,
			StartDate:          time.Now().UTC(),
			TotalInterest:      totalInterest,
			OutstandingBalance: outstandingBalance,
			PaymentsMade:       []model.Payment{},
		},
		LoanTermWeeks:  loanTermWeekly,
		WeeklyPayment:  weeklyPayment,
		WeeklyInterest: weeklyInterest,
	}

	// TODO: use storage dependency
	ls.Loans[tid.Suffix()] = loan

	return nil
}

// RecordPayment records a loan payment
func (ls *LoanService) RecordPayment(loanID model.LoanID, paymentAmount currency.Rupiah) error {
	loan, exists := ls.Loans[loanID.Suffix()]
	if !exists {
		return fmt.Errorf("loan not found")
	}

	// TODO: handle repayment
	// payment has to be exact with the weekly payment
	if loan.WeeklyPayment != paymentAmount {
		return fmt.Errorf("must pay exactly the same with the bill")
	}

	payment := model.Payment{
		Date:          time.Now().UTC(),
		Amount:        paymentAmount,
		BalanceBefore: loan.OutstandingBalance,
	}

	loan.OutstandingBalance = loan.OutstandingBalance.Subtract(paymentAmount)

	if paymentAmount >= loan.OutstandingBalance {
		payment.BalanceAfter = currency.NewRupiah(0, 0)
		payment.LoanCompleted = true
	} else {
		payment.BalanceAfter = loan.OutstandingBalance
	}

	loan.PaymentsMade = append(loan.PaymentsMade, payment)

	return nil
}

// GetNextPaymentDetails returns details for the next payment
func (ls *LoanService) GetNextPaymentDetails(loanID model.LoanID) (*model.Payment, error) {
	loan, exists := ls.Loans[loanID.Suffix()]
	if !exists {
		return nil, fmt.Errorf("loan not found")
	}

	var lastPaymentDate time.Time
	var paymentNumber int

	if len(loan.PaymentsMade) > 0 {
		lastPayment := loan.PaymentsMade[len(loan.PaymentsMade)-1]
		lastPaymentDate = lastPayment.Date
		paymentNumber = lastPayment.PaymentNumber + 1
	} else {
		lastPaymentDate = loan.StartDate
		paymentNumber = 1
	}

	return &model.Payment{
		Date:          lastPaymentDate.AddDate(0, 0, 7),
		PaymentNumber: paymentNumber,
		Amount:        loan.WeeklyPayment,
	}, nil
}

// CheckDelinquency checks loan delinquency status
func (ls *LoanService) CheckDelinquency(loanID model.LoanID) (*model.DelinquencyStatus, error) {
	loan, exists := ls.Loans[loanID.Suffix()]
	if !exists {
		return nil, fmt.Errorf("loan not found")
	}

	// No payments made yet
	if len(loan.PaymentsMade) == 0 {
		nextPayment, _ := ls.GetNextPaymentDetails(loanID)
		return &model.DelinquencyStatus{
			IsDelinquent:        false,
			MissedPayments:      0,
			LastPaymentDate:     loan.StartDate,
			NextExpectedPayment: nextPayment.Date,
			LateFee:             currency.NewRupiah(0, 0),
		}, nil
	}

	// get next expected payment
	nextPayment, _ := ls.GetNextPaymentDetails(loanID)

	// calculate missed payments
	missedPayments := 0
	lastPaymentDate := loan.PaymentsMade[len(loan.PaymentsMade)-1].Date

	// check for missed weekly payments
	for currentCheckDate := lastPaymentDate.AddDate(0, 0, 7); currentCheckDate.Before(time.Now()); currentCheckDate = currentCheckDate.AddDate(0, 0, 7) {
		missedPayments++
	}

	status := &model.DelinquencyStatus{
		IsDelinquent:        missedPayments >= 2,
		MissedPayments:      missedPayments,
		LastPaymentDate:     lastPaymentDate,
		NextExpectedPayment: nextPayment.Date,
		LateFee:             currency.NewRupiah(0, 0),
	}

	// calculate late fee if delinquent
	if status.IsDelinquent {
		// no late fee, itikad baik (0%)
		status.LateFee = loan.WeeklyPayment.Multiply(0).Divide(100)
	}

	return status, nil
}
