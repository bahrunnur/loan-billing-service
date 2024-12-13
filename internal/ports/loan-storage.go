package ports

import (
	"github.com/bahrunnur/loan-billing-service/internal/model"
)

type LoanCreator interface {
	CreateLoan(loan model.WeeklyLoan) error
}

type LoanGetter interface {
	GetLoan(loanID model.LoanID) (model.WeeklyLoan, error)
	GetLoanWithDelinquency(loanID model.LoanID) (model.WeeklyLoanWithDelinquency, error)
	GetLoanFullInformation(loanID model.LoanID) (model.WeeklyLoanFullInformation, error)
}

type LoanUpdater interface {
	UpdateLoan(loanID model.LoanID, updateParams model.WeeklyLoan) error
	UpdateLoanDelinquency(loanID model.LoanID, delinquency bool) error
}

type DelinquencyStatusCreator interface {
	CreateDelinquencyStatus(loanID model.LoanID, delinquencyStatus model.DelinquencyStatus) error
}

type DelinquencyStatusGetter interface {
	GetDelinquencyStatus(loanID model.LoanID) (model.DelinquencyStatus, error)
}

type DelinquencyStatusUpdater interface {
	UpdateDelinquencyStatus(loanID model.LoanID, updateParams model.DelinquencyStatus) error
}

type PaymentInserter interface {
	RecordPayment(loanID model.LoanID, payment model.Payment) error
}
