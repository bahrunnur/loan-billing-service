package memorystorage

import (
	"sync"
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/model"
)

type LoanStorage struct {
	// `mu` makes `MemoryStorage` to be thread-safe for parallel test
	mu                sync.RWMutex
	loans             map[model.LoanID]model.WeeklyLoan
	payments          map[model.LoanID][]model.Payment         // 1..n
	billings          map[model.LoanID][]model.Billing         // 1..n
	delinquencyStatus map[model.LoanID]model.DelinquencyStatus // 1..1
}

func NewLoanMemoryStorage() *LoanStorage {
	return &LoanStorage{
		loans:             map[model.LoanID]model.WeeklyLoan{},
		payments:          map[model.LoanID][]model.Payment{},
		billings:          map[model.LoanID][]model.Billing{},
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

func (ms *LoanStorage) CreateBilling(loanID model.LoanID, param model.BillingParam) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	currentDate := param.LoanCreationDate.UTC().AddDate(0, 0, 7)
	// emulate SQL COPY with Transaction block
	for i := range param.NumberOfTerm {
		b := model.Billing{
			LoanID:         loanID,
			TermNumber:     i + 1,
			Repayment:      param.RepaymentAmount,
			PaymentDueDate: currentDate,
			IsPaid:         false,
		}

		// stmt.Exec()
		ms.billings[loanID] = append(ms.billings[loanID], b)

		currentDate = currentDate.AddDate(0, 0, 7)
	}

	return nil
}

func (ms *LoanStorage) GetBillingAt(loanID model.LoanID, when time.Time) ([]model.Billing, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	billings, ok := ms.billings[loanID]
	if !ok {
		return nil, model.ErrLoanNotFound
	}

	ret := []model.Billing{}
	for _, b := range billings {
		if b.PaymentDueDate.Before(when.UTC()) {
			ret = append(ret, b)
		}
	}

	return ret, nil
}

func (ms *LoanStorage) GetUnfulfilledBillingAt(loanID model.LoanID, when time.Time) ([]model.Billing, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	billings, ok := ms.billings[loanID]
	if !ok {
		return nil, model.ErrLoanNotFound
	}

	padding := when.UTC().AddDate(0, 0, 7) // pad to a term

	// SQL WHERE due is before and not paid, sorted by due date

	ret := []model.Billing{}
	for _, b := range billings {
		if b.PaymentDueDate.Before(padding) && !b.IsPaid {
			ret = append(ret, b)
		}
	}

	return ret, nil
}

func (ms *LoanStorage) PayBillingUntil(loanID model.LoanID, when time.Time) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	billings, ok := ms.billings[loanID]
	if !ok {
		return model.ErrLoanNotFound
	}

	padding := when.UTC().AddDate(0, 0, 7) // pad to a term

	// SQL WHERE due is before and not paid, sorted by due date

	for i, b := range billings {
		if b.PaymentDueDate.Before(padding) && !b.IsPaid {
			billings[i].IsPaid = true
		}
	}

	ms.billings[loanID] = billings

	return nil
}
