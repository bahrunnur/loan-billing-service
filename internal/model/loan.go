package model

import (
	"time"

	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	"go.jetify.com/typeid"
)

type LoanPrefix struct{}

func (LoanPrefix) Prefix() string { return "loan" }

// Loan represents the structure of a loan
type Loan struct {
	ID                 typeid.TypeID[LoanPrefix] `json:"id"`
	Principal          currency.Rupiah           `json:"principal"`
	AnnualInterestRate int                       `json:"annual_interest_rate"` // basis point (1 basis point = 0.01%)
	StartDate          time.Time                 `json:"start_date"`
	TotalInterest      currency.Rupiah           `json:"total_interest"`
	OutstandingBalance currency.Rupiah           `json:"outstanding_balance"`
	PaymentsMade       []Payment                 `json:"payments_made"`
}

// WeeklyLoan is Loan for weekly term
type WeeklyLoan struct {
	Loan
	LoanTermWeeks int             `json:"loan_term_weeks"`
	WeeklyPayment currency.Rupiah `json:"weekly_payment"`
}

// WeeklyLoan is Loan for monthly term
type MonthlyLoan struct {
	Loan
	LoanTermMonths int             `json:"loan_term_months"`
	MonthlyPayment currency.Rupiah `json:"monthly_payment"`
}

// Payment represents a single loan payment
type Payment struct {
	Date          time.Time       `json:"date"`
	Amount        currency.Rupiah `json:"amount"`
	BalanceBefore currency.Rupiah `json:"balance_before"`
	BalanceAfter  currency.Rupiah `json:"balance_after"`
	LoanCompleted bool            `json:"loan_completed"`
}

// DelinquencyStatus represents the loan's delinquency details
type DelinquencyStatus struct {
	IsDelinquent bool            `json:"is_delinquent"`
	DaysOverdue  int             `json:"days_overdue"`
	LateFee      currency.Rupiah `json:"late_fee"`
}
