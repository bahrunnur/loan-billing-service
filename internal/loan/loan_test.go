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

	now := time.Now().UTC()

	testCases := []struct {
		name               string
		lastPaymentDate    time.Time
		checkDate          time.Time
		expectedDelinquent bool
		expectedError      error
	}{
		{
			name:               "Future Check Date",
			lastPaymentDate:    now.AddDate(0, 0, -4),
			checkDate:          now.AddDate(0, 0, 1),
			expectedDelinquent: false,
			expectedError:      model.ErrCheckFutureDelinquent,
		},
		{
			name:               "Not Delinquent - Recent Payment",
			lastPaymentDate:    now.AddDate(0, 0, -4),
			checkDate:          now,
			expectedDelinquent: false,
			expectedError:      nil,
		},
		{
			name:               "Not Delinquent - Only Missed a Threshold",
			lastPaymentDate:    now.AddDate(0, 0, -((7 * model.MISSED_PAYMENT_THRESHOLD) + 1)), // 8 days
			checkDate:          now,
			expectedDelinquent: false,
			expectedError:      nil,
		},
		{
			name:               "Not Delinquent - Exactly at Threshold",
			lastPaymentDate:    now.AddDate(0, 0, -(7 * (model.MISSED_PAYMENT_THRESHOLD + 1))), // 14 days
			checkDate:          now,
			expectedDelinquent: false,
			expectedError:      nil,
		},
		{
			name:               "Delinquent - Beyond Threshold",
			lastPaymentDate:    now.AddDate(0, 0, -((7 * (model.MISSED_PAYMENT_THRESHOLD + 1)) + 1)), // 15 days
			checkDate:          now,
			expectedDelinquent: true,
			expectedError:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := memorystorage.NewLoanMemoryStorage()
			loanService := loan.NewLoanService(memStorage)

			loanID, err := typeid.New[model.LoanID]()
			g.Expect(err).ToNot(HaveOccurred())

			loan := createBareboneLoan(loanID, now)
			delinquency := createBareboneDelinquency(loanID, now)
			delinquency.LastPaymentDate = tc.lastPaymentDate

			err = memStorage.CreateLoan(loan)
			g.Expect(err).ToNot(HaveOccurred())

			err = memStorage.CreateDelinquencyStatus(loanID, delinquency)
			g.Expect(err).ToNot(HaveOccurred())

			isDelinquent, err := loanService.CheckDelinquency(loan.ID, tc.checkDate)

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
		name                    string
		paymentMultiplier       int
		lastPaymentDate         time.Time
		NextExpectedPaymentDate time.Time
		currentPaymentDate      time.Time
		expectedError           error
		expectedOutstanding     currency.Rupiah
	}{
		{
			name:                    "Successful - On-Time Payment",
			paymentMultiplier:       1,
			lastPaymentDate:         now.AddDate(0, 0, -4),
			NextExpectedPaymentDate: now.AddDate(0, 0, 3),
			currentPaymentDate:      now,
			expectedError:           nil,
			expectedOutstanding:     currency.NewRupiah(5500000-110000, 0),
		},
		{
			name:                    "Succesful - Missed 1 Payment",
			paymentMultiplier:       2,
			lastPaymentDate:         now.AddDate(0, 0, -((7 * model.MISSED_PAYMENT_THRESHOLD) + 1)), // 8 days
			NextExpectedPaymentDate: now.AddDate(0, 0, -7),
			currentPaymentDate:      now,
			expectedError:           nil,
			expectedOutstanding:     currency.NewRupiah(5500000-(110000*2), 0),
		},
		{
			name:                    "Fail - Pay All Term",
			paymentMultiplier:       weeklyLoanTerm,
			lastPaymentDate:         now.AddDate(0, 0, -4),
			NextExpectedPaymentDate: now.AddDate(0, 0, 3),
			currentPaymentDate:      now,
			expectedError:           model.ErrMismatchPayment,
		},
		{
			name:                    "Fail - Missed 1 Payment - Payment is less than needed",
			paymentMultiplier:       1,
			lastPaymentDate:         now.AddDate(0, 0, -((7 * model.MISSED_PAYMENT_THRESHOLD) + 1)), // 8 days
			NextExpectedPaymentDate: now.AddDate(0, 0, -7),
			currentPaymentDate:      now,
			expectedError:           model.ErrMismatchPayment,
		},
		{
			name:                    "Fail - Missed 1 Payment - Payment is more than needed",
			paymentMultiplier:       3,
			lastPaymentDate:         now.AddDate(0, 0, -((7 * model.MISSED_PAYMENT_THRESHOLD) + 1)), // 8 days
			NextExpectedPaymentDate: now.AddDate(0, 0, -7),
			currentPaymentDate:      now,
			expectedError:           model.ErrMismatchPayment,
		},
		{
			name:                    "Fail - Missed 2 Payments - Flagged as Delinquent",
			paymentMultiplier:       3,
			lastPaymentDate:         now.AddDate(0, 0, -((7 * (model.MISSED_PAYMENT_THRESHOLD + 1)) + 1)), // 15 days
			NextExpectedPaymentDate: now.AddDate(0, 0, -7),
			currentPaymentDate:      now,
			expectedError:           model.ErrPayInDelinquent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := memorystorage.NewLoanMemoryStorage()
			loanService := loan.NewLoanService(memStorage)

			// create a loan first
			loan, err := loanService.CreateLoan(principal, interestRate, weeklyLoanTerm)
			g.Expect(err).ToNot(HaveOccurred())

			// update last payment
			updatedStatus, err := memStorage.GetDelinquencyStatus(loan.ID)
			g.Expect(err).ToNot(HaveOccurred())
			updatedStatus.LastPaymentDate = tc.lastPaymentDate
			updatedStatus.NextExpectedPaymentDate = tc.NextExpectedPaymentDate
			err = memStorage.UpdateDelinquencyStatus(loan.ID, updatedStatus)
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

func createBareboneLoan(loanID model.LoanID, now time.Time) model.WeeklyLoan {
	return model.WeeklyLoan{
		Loan: model.Loan{
			ID:                 loanID,
			Principal:          currency.NewRupiah(0, 0),
			AnnualInterestRate: model.BPS(1000),
			StartDate:          now.UTC(),
			TotalInterest:      currency.NewRupiah(0, 0),
			OutstandingBalance: currency.NewRupiah(0, 0),
		},
		LoanTermWeeks:  10,
		WeeklyPayment:  currency.NewRupiah(0, 0),
		WeeklyInterest: currency.NewRupiah(0, 0),
	}
}

func createBareboneDelinquency(loanID model.LoanID, now time.Time) model.DelinquencyStatus {
	return model.DelinquencyStatus{
		LoanID:                  loanID,
		IsDelinquent:            false,
		LastPaymentDate:         now.UTC(),
		NextExpectedPaymentDate: now.UTC().AddDate(0, 0, 7),
		LateFee:                 currency.NewRupiah(0, 0),
	}
}
