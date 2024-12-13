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

func NewLoanMemoryStorage() *LoanStorage {
	return &LoanStorage{
		loans:             map[model.LoanID]model.WeeklyLoan{},
		payments:          map[model.LoanID][]model.Payment{},
		delinquencyStatus: map[model.LoanID]model.DelinquencyStatus{},
	}
}

func (ms *LoanStorage) CreateLoan(loan model.WeeklyLoan) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.loans[loan.ID] = loan

	return nil
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

func (ms *LoanStorage) GetLoanWithDelinquency(loanID model.LoanID) (model.WeeklyLoanWithDelinquency, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	loan, ok := ms.loans[loanID]
	if !ok {
		return model.WeeklyLoanWithDelinquency{}, model.ErrLoanNotFound
	}

	delinquency, ok := ms.delinquencyStatus[loanID]
	if !ok {
		return model.WeeklyLoanWithDelinquency{}, model.ErrDelinquencyStatusNotFound
	}

	// emulate SQL JOIN
	ret := model.WeeklyLoanWithDelinquency{
		WeeklyLoan:        loan,
		DelinquencyStatus: delinquency,
	}

	return ret, nil
}

func (ms *LoanStorage) GetLoanFullInformation(loanID model.LoanID) (model.WeeklyLoanFullInformation, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	loan, ok := ms.loans[loanID]
	if !ok {
		return model.WeeklyLoanFullInformation{}, model.ErrLoanNotFound
	}

	delinquency, ok := ms.delinquencyStatus[loanID]
	if !ok {
		return model.WeeklyLoanFullInformation{}, model.ErrDelinquencyStatusNotFound
	}

	payments := ms.payments[loanID]

	// emulate SQL JOIN
	ret := model.WeeklyLoanFullInformation{
		WeeklyLoan:        loan,
		DelinquencyStatus: delinquency,
		Payments:          payments,
	}

	return ret, nil
}

func (ms *LoanStorage) UpdateLoan(loanID model.LoanID, updateParams model.WeeklyLoan) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.loans[loanID] = updateParams

	return nil
}

func (ms *LoanStorage) UpdateLoanDelinquency(loanID model.LoanID, delinquency bool) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	d, ok := ms.delinquencyStatus[loanID]
	if !ok {
		return model.ErrLoanNotFound
	}

	// emulate SQL UPDATE on 1 column
	d.IsDelinquent = delinquency
	ms.delinquencyStatus[loanID] = d

	return nil
}

func (ms *LoanStorage) CreateDelinquencyStatus(loanID model.LoanID, delinquencyStatus model.DelinquencyStatus) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.delinquencyStatus[loanID] = delinquencyStatus

	return nil
}

func (ms *LoanStorage) GetDelinquencyStatus(loanID model.LoanID) (model.DelinquencyStatus, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delinquencyStatus, ok := ms.delinquencyStatus[loanID]
	if !ok {
		return model.DelinquencyStatus{}, model.ErrLoanNotFound
	}

	return delinquencyStatus, nil
}

func (ms *LoanStorage) UpdateDelinquencyStatus(loanID model.LoanID, updateParams model.DelinquencyStatus) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.delinquencyStatus[loanID] = updateParams

	return nil
}

func (ms *LoanStorage) RecordPayment(loanID model.LoanID, payment model.Payment) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.payments[loanID] = append(ms.payments[loanID], payment)

	return nil
}
