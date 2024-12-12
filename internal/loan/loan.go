package loan

import (
	"github.com/bahrunnur/loan-billing-service/internal/model"
	"go.jetify.com/typeid"
)

// LoanService manages loan-related operations
type LoanService struct {
	loans map[typeid.TypeID[model.LoanPrefix]]*model.Loan
}

// NewLoanService creates a new LoanService
func NewLoanService() *LoanService {
	return &LoanService{
		loans: make(map[typeid.TypeID[model.LoanPrefix]]*model.Loan),
	}
}
