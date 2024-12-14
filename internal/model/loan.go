package model

import (
	"time"

	"github.com/bahrunnur/loan-billing-service/pkg/currency"
)

// Loan represents the structure of a loan
type Loan struct {
	ID                 LoanID          `json:"id"`
	Principal          currency.Rupiah `json:"principal"`
	AnnualInterestRate BPS             `json:"annual_interest_rate"` // basis point (1 basis point = 0.01%)
	StartDate          time.Time       `json:"start_date"`
	TotalInterest      currency.Rupiah `json:"total_interest"`
	OutstandingBalance currency.Rupiah `json:"outstanding_balance"`
	IsCompleted        bool            `json:"is_completed"`
}

// WeeklyLoan is Loan for weekly term
type WeeklyLoan struct {
	Loan
	LoanTermWeeks  int             `json:"loan_term_weeks"`
	WeeklyPayment  currency.Rupiah `json:"weekly_payment"`
	WeeklyInterest currency.Rupiah `json:"weekly_interest"`
}

// MonthlyLoan is Loan for monthly term, not used at the moment
type MonthlyLoan struct {
	Loan
	LoanTermMonths int             `json:"loan_term_months"`
	MonthlyPayment currency.Rupiah `json:"monthly_payment"`
}

// Payment represents a single loan payment
type Payment struct {
	LoanID        LoanID          `json:"loan_id"`
	Date          time.Time       `json:"date"`
	Amount        currency.Rupiah `json:"amount"`
	BalanceBefore currency.Rupiah `json:"balance_before"`
	BalanceAfter  currency.Rupiah `json:"balance_after"`
}

type Billing struct {
	LoanID         LoanID          `json:"loan_id"`
	TermNumber     int             `json:"term_number"`
	PaymentDueDate time.Time       `json:"payment_due_date"`
	Repayment      currency.Rupiah `json:"repayment"`
	IsPaid         bool            `json:"is_paid"`
}

type BillingParam struct {
	LoanCreationDate time.Time
	NumberOfTerm     int
	RepaymentAmount  currency.Rupiah
}

// DelinquencyStatus represents the loan's delinquency details
type DelinquencyStatus struct {
	LoanID                  LoanID          `json:"loan_id"`
	IsDelinquent            bool            `json:"is_delinquent"`
	LastPaymentDate         time.Time       `json:"last_payment_date"`
	NextExpectedPaymentDate time.Time       `json:"next_expected_payment_date"`
	LateFee                 currency.Rupiah `json:"late_fee"`
}

type WeeklyLoanWithDelinquency struct {
	WeeklyLoan
	DelinquencyStatus
}

type WeeklyLoanFullInformation struct {
	WeeklyLoan
	DelinquencyStatus
	Payments []Payment `json:"payments"`
}
