package model

import "errors"

var (
	ErrLoanNotFound              = errors.New("loan not found")
	ErrPaymentNotFound           = errors.New("payment not found")
	ErrDelinquencyStatusNotFound = errors.New("delinquency status not found")

	ErrNegativeInterest      = errors.New("expect a positive interest")
	ErrNoPrincipal           = errors.New("expect some principal")
	ErrNoTerm                = errors.New("expect a term")
	ErrMismatchPayment       = errors.New("expect a correct exact payment")
	ErrPayInDelinquent       = errors.New("expect not delinquent")
	ErrCheckFutureDelinquent = errors.New("expect past and present")
	ErrRepaymentComplete     = errors.New("expect loan not complete")
)
