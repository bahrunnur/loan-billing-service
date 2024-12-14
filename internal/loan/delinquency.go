package loan

import (
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/model"
)

// ColdDelinquentFlag do a search through db to check the account delinquency
func (ls *LoanService) ColdDelinquentFlag(loanID model.LoanID, checkAt time.Time) (bool, []model.Billing, error) {
	// I assume the account is delinquent after missing payment 2 times,
	// and no repayment have been made before the week #2 due date

	unfulfilledBilling, err := ls.storage.GetUnfulfilledBillingAt(loanID, checkAt.UTC())
	if err != nil {
		return false, nil, err
	}

	if len(unfulfilledBilling) > model.MISSED_PAYMENT_THRESHOLD+1 {
		return true, unfulfilledBilling, nil
	}

	return false, unfulfilledBilling, nil
}
