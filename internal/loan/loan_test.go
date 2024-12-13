package loan_test

import (
	"testing"
	"time"

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
		expectedError      bool
	}{
		{
			name:               "Valid Loan Creation",
			principal:          currency.NewRupiah(1000000, 0),
			annualInterestRate: model.BPS(1000), // 10%
			loanTermWeekly:     10,
			expectedError:      false,
		},
		{
			name:               "Zero Principal",
			principal:          currency.NewRupiah(0, 0),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     10,
			expectedError:      true,
		},
		{
			name:               "Zero Rupiah Principal",
			principal:          currency.NewRupiah(0, 50),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     10,
			expectedError:      false,
		},
		{
			name:               "Zero Sen Principal",
			principal:          currency.NewRupiah(500000, 0),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     10,
			expectedError:      false,
		},
		{
			name:               "Zero Loan Term",
			principal:          currency.NewRupiah(1000000, 0),
			annualInterestRate: model.BPS(1000),
			loanTermWeekly:     0,
			expectedError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loanService := loan.NewLoanService()
			loan, err := loanService.CreateLoan(tc.principal, tc.annualInterestRate, tc.loanTermWeekly)

			if tc.expectedError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(loan).To(BeNil())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(loanService.Loans).To(HaveLen(1))
				g.Expect(loan).ToNot(BeNil())
			}
		})
	}
}

func TestRecordPayment(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	loanService := loan.NewLoanService()

	// create a loan first
	principal := currency.NewRupiah(1000000, 0)
	interestRate := model.BPS(1000)
	loanTermWeekly := 10
	loan, err := loanService.CreateLoan(principal, interestRate, loanTermWeekly)
	g.Expect(err).ToNot(HaveOccurred())

	// get the created loan ID
	loanID := loan.ID

	testCases := []struct {
		name          string
		paymentAmount currency.Rupiah
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "Correct Weekly Payment",
			paymentAmount: loanService.Loans[loanID.Suffix()].WeeklyPayment,
			expectedError: false,
		},
		{
			name:          "Incorrect Payment Amount",
			paymentAmount: currency.NewRupiah(50000, 0),
			expectedError: true,
			errorMessage:  "must pay exactly the same with the bill",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := loanService.RecordPayment(loanID, tc.paymentAmount)

			if tc.expectedError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(loanService.Loans[loanID.Suffix()].PaymentsMade).ToNot(BeEmpty())
			}
		})
	}
}

func TestGetLoan(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	loanService := loan.NewLoanService()

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
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "Loan Found",
			loanID:        createdLoan.ID,
			expectedError: false,
		},
		{
			name:          "Loan Not Found with Random ID",
			loanID:        randoID,
			expectedError: true,
			errorMessage:  model.ErrLoanNotFound.Error(),
		},
		{
			name:          "Loan Not Found with Zero ID",
			loanID:        zeroID,
			expectedError: true,
			errorMessage:  model.ErrLoanNotFound.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loan, err := loanService.GetLoan(tc.loanID)

			if tc.expectedError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal(tc.errorMessage))
				g.Expect(loan).To(BeNil())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(loan).ToNot(BeNil())
				g.Expect(loan.ID).To(Equal(tc.loanID))
				g.Expect(loan.OutstandingBalance).To(Equal(currency.NewRupiah(1100000, 00)))
			}
		})
	}
}

func TestGetNextPaymentDetails(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	loanService := loan.NewLoanService()

	// create a loan first
	principal := currency.NewRupiah(1000000, 0)
	interestRate := model.BPS(1000)
	loanTermWeekly := 10
	loan, err := loanService.CreateLoan(principal, interestRate, loanTermWeekly)
	g.Expect(err).ToNot(HaveOccurred())

	// get the created loan ID
	loanID := loan.ID

	t.Run("First Payment Details", func(t *testing.T) {
		nextPayment, err := loanService.GetNextPaymentDetails(loanID)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(nextPayment).ToNot(BeNil())
		g.Expect(nextPayment.PaymentNumber).To(Equal(1))
		g.Expect(loanService.Loans[loanID.Suffix()].WeeklyPayment).To(Equal(nextPayment.Amount))
	})

	// record a payment and check next payment details
	t.Run("Next Payment Details After First Payment", func(t *testing.T) {
		weeklyPayment := loanService.Loans[loanID.Suffix()].WeeklyPayment
		err := loanService.RecordPayment(loanID, weeklyPayment)
		g.Expect(err).ToNot(HaveOccurred())

		nextPayment, err := loanService.GetNextPaymentDetails(loanID)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(nextPayment).ToNot(BeNil())
		g.Expect(nextPayment.PaymentNumber).To(Equal(1))
	})
}

func TestCheckDelinquency(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	loanService := loan.NewLoanService()

	// create a loan first
	principal := currency.NewRupiah(1000000, 0)
	interestRate := model.BPS(1000)
	loanTermWeekly := 10
	loan, err := loanService.CreateLoan(principal, interestRate, loanTermWeekly)
	g.Expect(err).ToNot(HaveOccurred())

	// get the created loan ID
	loanID := loan.ID

	t.Run("No Payments Made", func(t *testing.T) {
		delinquencyStatus, err := loanService.CheckDelinquency(loanID)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(delinquencyStatus.IsDelinquent).To(BeFalse())
		g.Expect(delinquencyStatus.MissedPayments).To(BeZero())
	})

	t.Run("Delinquency After Multiple Missed Payments", func(t *testing.T) {
		// simulate time passing without payments
		weeklyPayment := loanService.Loans[loanID.Suffix()].WeeklyPayment

		// manually modify the last payment date to be in the past
		loan := loanService.Loans[loanID.Suffix()]
		oldestPossibleDate := time.Now().AddDate(0, 0, -14) // 2 weeks back
		loan.PaymentsMade = []model.Payment{
			{
				Date:   oldestPossibleDate,
				Amount: weeklyPayment,
			},
		}

		delinquencyStatus, err := loanService.CheckDelinquency(loanID)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(delinquencyStatus.IsDelinquent).To(BeTrue())
		g.Expect(delinquencyStatus.MissedPayments).To(Equal(2))
	})
}

func TestLoanNotFound(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	loanService := loan.NewLoanService()
	nonExistentLoanID, err := typeid.New[model.LoanID]()
	g.Expect(err).ToNot(HaveOccurred())

	testCases := []struct {
		name     string
		testFunc func() error
		errorMsg string
	}{
		{
			name: "RecordPayment with Non-Existent Loan",
			testFunc: func() error {
				return loanService.RecordPayment(nonExistentLoanID, currency.NewRupiah(1000, 0))
			},
			errorMsg: "loan not found",
		},
		{
			name: "GetNextPaymentDetails with Non-Existent Loan",
			testFunc: func() error {
				_, err := loanService.GetNextPaymentDetails(nonExistentLoanID)
				return err
			},
			errorMsg: "loan not found",
		},
		{
			name: "CheckDelinquency with Non-Existent Loan",
			testFunc: func() error {
				_, err := loanService.CheckDelinquency(nonExistentLoanID)
				return err
			},
			errorMsg: "loan not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			g.Expect(err.Error()).To(Equal(tc.errorMsg))
		})
	}
}
