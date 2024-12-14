package loan_test

import (
	"testing"
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/adapters/memorystorage"
	"github.com/bahrunnur/loan-billing-service/internal/loan"
	"github.com/bahrunnur/loan-billing-service/internal/model"
	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	. "github.com/onsi/gomega"
	"go.jetify.com/typeid"
)

func TestCreateLoan(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	testCases := []struct {
		name               string
		principal          currency.Rupiah
		annualInterestRate model.BPS
		loanTermWeekly     int
		expectedError      error
	}{
		{
			name:               "Valid Loan Creation",
			principal:          currency.NewRupiah(1000000, 0),
			annualInterestRate: model.BPS(1000), // 10%
			loanTermWeekly:     10,
			expectedError:      nil,
		},
		{
			name:               "Negative Interest Rate",
			principal:          currency.NewRupiah(1000000, 0),
			annualInterestRate: model.BPS(-100),
			loanTermWeekly:     10,
			expectedError:      model.ErrNegativeInterest,
		},
		{
			name:               "Zero Principal",
			principal:          currency.NewRupiah(0, 0),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     10,
			expectedError:      model.ErrNoPrincipal,
		},
		{
			name:               "Zero Rupiah Principal",
			principal:          currency.NewRupiah(0, 50),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     10,
			expectedError:      nil,
		},
		{
			name:               "Zero Sen Principal",
			principal:          currency.NewRupiah(500000, 0),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     10,
			expectedError:      nil,
		},
		{
			name:               "Zero Loan Term",
			principal:          currency.NewRupiah(1000000, 0),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     0,
			expectedError:      model.ErrNoTerm,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := memorystorage.NewLoanMemoryStorage()
			loanService := loan.NewLoanService(memStorage)
			createdLoan, err := loanService.CreateLoan(tc.principal, tc.annualInterestRate, tc.loanTermWeekly)

			if tc.expectedError != nil {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(tc.expectedError))
				g.Expect(createdLoan).To(BeZero())
			} else {
				g.Expect(err).ToNot(HaveOccurred())

				actual, err := memStorage.GetLoan(createdLoan.ID)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(createdLoan).ToNot(BeNil())
				g.Expect(createdLoan).To(Equal(actual))
			}
		})
	}
}

func TestCheckDelinquency(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	memStorage := memorystorage.NewLoanMemoryStorage()
	loanService := loan.NewLoanService(memStorage)

	now := time.Now().UTC()

	// create a loan first
	principal := currency.NewRupiah(1000000, 0)
	interestRate := model.BPS(1000)
	loanTermWeekly := 10
	createdLoan, err := loanService.CreateLoan(principal, interestRate, loanTermWeekly)
	g.Expect(err).ToNot(HaveOccurred())

	testCases := []struct {
		name               string
		checkDate          time.Time
		expectedDelinquent bool
		expectedError      error
	}{
		{
			name:               "Not Delinquent - Loan just Created",
			checkDate:          now.AddDate(0, 0, 1),
			expectedDelinquent: false,
			expectedError:      nil,
		},
		{
			name:               "Not Delinquent - Only Missed a Threshold",
			checkDate:          now.AddDate(0, 0, (7*model.MISSED_PAYMENT_THRESHOLD)+1), // 8 days,,
			expectedDelinquent: false,
			expectedError:      nil,
		},
		{
			name:               "Not Delinquent - Exactly at Threshold",
			checkDate:          now.AddDate(0, 0, 7*(model.MISSED_PAYMENT_THRESHOLD+1)), // 14 days
			expectedDelinquent: false,
			expectedError:      nil,
		},
		{
			name:               "Delinquent - Beyond Threshold",
			checkDate:          now.AddDate(0, 0, ((7 * (model.MISSED_PAYMENT_THRESHOLD + 1)) + 1)), // 15 days
			expectedDelinquent: true,
			expectedError:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDelinquent, err := loanService.CheckDelinquency(createdLoan.ID, tc.checkDate)

			if tc.expectedError != nil {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(tc.expectedError))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(isDelinquent).To(Equal(tc.expectedDelinquent))
			}
		})
	}
}

func TestRecordPayment(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	now := time.Now().UTC()
	principal := currency.NewRupiah(5000000, 0)
	interestRate := model.BPS(1000)
	weeklyLoanTerm := 50

	testCases := []struct {
		name                string
		paymentMultiplier   int
		currentPaymentDate  time.Time
		expectedError       error
		expectedOutstanding currency.Rupiah
	}{
		{
			name:                "Successful - On-Time Payment",
			paymentMultiplier:   1,
			currentPaymentDate:  now.AddDate(0, 0, 2),
			expectedError:       nil,
			expectedOutstanding: currency.NewRupiah(5500000-110000, 0),
		},
		{
			name:                "Succesful - Missed a Payment",
			paymentMultiplier:   2,
			currentPaymentDate:  now.AddDate(0, 0, (7*model.MISSED_PAYMENT_THRESHOLD)+1), // 8 days
			expectedError:       nil,
			expectedOutstanding: currency.NewRupiah(5500000-(110000*2), 0),
		},
		{
			name:               "Fail - Pay All Term",
			paymentMultiplier:  weeklyLoanTerm,
			currentPaymentDate: now.AddDate(0, 0, 2),
			expectedError:      model.ErrMismatchPayment,
		},
		{
			name:               "Fail - Missed 1 Payment - Payment is less than needed",
			paymentMultiplier:  1,
			currentPaymentDate: now.AddDate(0, 0, (7*model.MISSED_PAYMENT_THRESHOLD)+1), // 8 days
			expectedError:      model.ErrMismatchPayment,
		},
		{
			name:               "Fail - Missed 1 Payment - Payment is more than needed",
			paymentMultiplier:  3,
			currentPaymentDate: now.AddDate(0, 0, (7*model.MISSED_PAYMENT_THRESHOLD)+1), // 8 days
			expectedError:      model.ErrMismatchPayment,
		},
		{
			name:               "Fail - Missed 2 Payments - Flagged as Delinquent",
			paymentMultiplier:  3,
			currentPaymentDate: now.AddDate(0, 0, ((7 * (model.MISSED_PAYMENT_THRESHOLD + 1)) + 1)), // 15 days
			expectedError:      model.ErrPayInDelinquent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := memorystorage.NewLoanMemoryStorage()
			loanService := loan.NewLoanService(memStorage)

			// create a loan first
			loan, err := loanService.CreateLoan(principal, interestRate, weeklyLoanTerm)
			g.Expect(err).ToNot(HaveOccurred())

			err = loanService.RecordPayment(loan.ID, tc.currentPaymentDate, loan.WeeklyPayment.Multiply(tc.paymentMultiplier))

			if tc.expectedError != nil {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(tc.expectedError))
			} else {
				g.Expect(err).ToNot(HaveOccurred())

				updatedLoan, err := loanService.GetLoan(loan.ID)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(updatedLoan.OutstandingBalance).To(Equal(tc.expectedOutstanding))
			}
		})
	}
}

func TestGetLoan(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	memStorage := memorystorage.NewLoanMemoryStorage()
	loanService := loan.NewLoanService(memStorage)

	// create a loan first
	principal := currency.NewRupiah(1000000, 0)
	interestRate := model.BPS(1000)
	loanTermWeekly := 10
	createdLoan, err := loanService.CreateLoan(principal, interestRate, loanTermWeekly)
	g.Expect(err).ToNot(HaveOccurred())

	randoID, err := typeid.New[model.LoanID]()
	g.Expect(err).ToNot(HaveOccurred())

	zeroID, err := typeid.FromSuffix[model.LoanID]("00000000000000000000000000")
	g.Expect(err).ToNot(HaveOccurred())

	testCases := []struct {
		name          string
		loanID        model.LoanID
		expectedError error
	}{
		{
			name:          "Loan Found",
			loanID:        createdLoan.ID,
			expectedError: nil,
		},
		{
			name:          "Loan Not Found with Random ID",
			loanID:        randoID,
			expectedError: model.ErrLoanNotFound,
		},
		{
			name:          "Loan Not Found with Zero ID",
			loanID:        zeroID,
			expectedError: model.ErrLoanNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loan, err := loanService.GetLoan(tc.loanID)

			if tc.expectedError != nil {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(tc.expectedError))
				g.Expect(loan).To(BeZero())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(loan).ToNot(BeNil())
				g.Expect(loan.ID).To(Equal(tc.loanID))
				g.Expect(loan.OutstandingBalance).To(Equal(currency.NewRupiah(1100000, 00)))
			}
		})
	}
}

func TestColdDelinquentFlag(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	memStorage := memorystorage.NewLoanMemoryStorage()
	loanService := loan.NewLoanService(memStorage)

	now := time.Now().UTC()

	// create a loan first
	principal := currency.NewRupiah(1000000, 0)
	interestRate := model.BPS(1000)
	loanTermWeekly := 10
	createdLoan, err := loanService.CreateLoan(principal, interestRate, loanTermWeekly)
	g.Expect(err).ToNot(HaveOccurred())

	randoID, err := typeid.New[model.LoanID]()
	g.Expect(err).ToNot(HaveOccurred())

	testCases := []struct {
		name                             string
		loanID                           model.LoanID
		checkAt                          time.Time
		expectedDelinquent               bool
		expectedUnfulfilledBillingLength int
		expectedError                    error
	}{
		{
			name:                             "Not Delinquent",
			loanID:                           createdLoan.ID,
			checkAt:                          now.AddDate(0, 0, 2),
			expectedDelinquent:               false,
			expectedUnfulfilledBillingLength: 1,
			expectedError:                    nil,
		},
		{
			name:                             "Not Delinquent - Missed a Payment",
			loanID:                           createdLoan.ID,
			checkAt:                          now.AddDate(0, 0, (7*model.MISSED_PAYMENT_THRESHOLD)+1), // 8 days
			expectedDelinquent:               false,
			expectedUnfulfilledBillingLength: 2,
			expectedError:                    nil,
		},
		{
			name:                             "Delinquent - Haven't Made a Payment Beyond Threshold",
			loanID:                           createdLoan.ID,
			checkAt:                          now.AddDate(0, 0, ((7 * (model.MISSED_PAYMENT_THRESHOLD + 1)) + 1)), // 15 days
			expectedDelinquent:               true,
			expectedUnfulfilledBillingLength: 3,
			expectedError:                    nil,
		},
		{
			name:                             "Not Delinquent - Loan Not Found",
			loanID:                           randoID,
			checkAt:                          now.AddDate(0, 0, 2),
			expectedDelinquent:               false,
			expectedUnfulfilledBillingLength: 0,
			expectedError:                    model.ErrLoanNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDelinquent, unfulfilledBilling, err := loanService.ColdDelinquentFlag(tc.loanID, tc.checkAt)
			g.Expect(isDelinquent).To(Equal(tc.expectedDelinquent))
			g.Expect(unfulfilledBilling).To(HaveLen(tc.expectedUnfulfilledBillingLength))

			if tc.expectedError != nil {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(tc.expectedError))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
