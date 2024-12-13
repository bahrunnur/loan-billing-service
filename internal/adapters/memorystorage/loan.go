package memorystorage

import (
	"sync"

	"github.com/bahrunnur/loan-billing-service/internal/model"
)

type LoanStorage struct {
	// `mu` makes `MemoryStorage` to be thread-safe for parallel test
	mu                sync.RWMutex
	loans             map[model.LoanID]model.WeeklyLoan
	payments          map[model.LoanID][]model.Payment         // 1..n
	delinquencyStatus map[model.LoanID]model.DelinquencyStatus // 1..1
}

func NewLoanStorage() *LoanStorage {
	return &LoanStorage{
		loans:             map[model.LoanID]model.WeeklyLoan{},
		payments:          map[model.LoanID][]model.Payment{},
		delinquencyStatus: map[model.LoanID]model.DelinquencyStatus{},
	}
}

func (ms *LoanStorage) GetLoan(loanID model.LoanID) (model.WeeklyLoan, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	loan, ok := ms.loans[loanID]
	if !ok {
		return model.WeeklyLoan{}, model.ErrLoanNotFound
	}

	return loan, nil
}

func (ms *LoanStorage) CreateLoan(loan model.WeeklyLoan) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.loans[loan.ID] = loan

	return nil
}
