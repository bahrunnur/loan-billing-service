package loan

import (
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/model"
)

// isDelinquent return true if no payment has been made to checkAt time
func (ls *LoanService) isDelinquent(lastPayment, checkAt time.Time) bool {
	lastPayment = lastPayment.UTC()
	checkAt = checkAt.UTC()

	// I assume the account is delinquent after missing payment 2 times,
	// and no repayment have been made before the week #2 due date
	return checkAt.After(lastPayment.AddDate(0, 0, 7*(model.MISSED_PAYMENT_THRESHOLD+1)))
}
