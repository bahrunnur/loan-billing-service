package model

import (
	"time"

	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	"go.jetify.com/typeid"
)

type LoanPrefix struct{}

func (LoanPrefix) Prefix() string { return "loan" }

type LoanID struct {
	typeid.TypeID[LoanPrefix]
}

// Loan represents the structure of a loan
type Loan struct {
	ID                 LoanID          `json:"id"`
	Principal          currency.Rupiah `json:"principal"`
	AnnualInterestRate BPS             `json:"annual_interest_rate"` // basis point (1 basis point = 0.01%)
	StartDate          time.Time       `json:"start_date"`
	TotalInterest      currency.Rupiah `json:"total_interest"`
	OutstandingBalance currency.Rupiah `json:"outstanding_balance"`
	PaymentsMade       []Payment       `json:"payments_made"`
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
	PaymentNumber int             `json:"payment_number"`
	Amount        currency.Rupiah `json:"amount"`
	BalanceBefore currency.Rupiah `json:"balance_before"`
	BalanceAfter  currency.Rupiah `json:"balance_after"`
	LoanCompleted bool            `json:"loan_completed"`
}

// DelinquencyStatus represents the loan's delinquency details
type DelinquencyStatus struct {
	LoanID              LoanID          `json:"loan_id"`
	IsDelinquent        bool            `json:"is_delinquent"`
	MissedPayments      int             `json:"missed_payments"`
	LastPaymentDate     time.Time       `json:"last_payment_date"`
	NextExpectedPayment time.Time       `json:"next_expected_payment"`
	LateFee             currency.Rupiah `json:"late_fee"`
}

const PERCENT = 100

// BPS is a basis point
type BPS int

func (b BPS) ToPercentage() int { // TODO: use a better way like math/big
	return int(b) / 100
}

func FromPercentage(p int) BPS {
	return BPS(p * 100)
}
