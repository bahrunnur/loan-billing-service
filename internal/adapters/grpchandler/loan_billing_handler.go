package grpchandler

import (
	"context"
	"log"
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/model"
	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	"github.com/bahrunnur/loan-billing-service/pkg/o11y"
	v1 "github.com/bahrunnur/loan-billing-service/proto/gen/loanbilling/v1"
	"go.jetify.com/typeid"
	"go.uber.org/zap"
)

type LoanBillingService interface {
	GetLoan(loanID model.LoanID) (model.WeeklyLoanFullInformation, error)
	CheckDelinquency(loanID model.LoanID, when time.Time) (bool, error)
	RecordPayment(loanID model.LoanID, when time.Time, paymentAmount currency.Rupiah) error
}

type LoanBillingGRPCServer struct {
	svc LoanBillingService
	v1.UnimplementedLoanBillingServiceServer
}

func NewLoanBillingGRPCServer(svc LoanBillingService) *LoanBillingGRPCServer {
	return &LoanBillingGRPCServer{svc: svc}
}

func (s *LoanBillingGRPCServer) GetOutstanding(ctx context.Context, req *v1.GetOutstandingRequest) (*v1.GetOutstandingResponse, error) {
	logger, ok := ctx.Value(o11y.LoggerKey{}).(*zap.Logger)
	if !ok {
		log.Println("no logger in ctx, fallback to no-op logger")
		logger = zap.NewNop()
	}

	loanID, err := typeid.Parse[model.LoanID](req.LoanId)
	if err != nil {
		logger.Error("fail to parse loan id",
			zap.String("requested_loan_id", req.LoanId),
		)
		return nil, err
	}

	loan, err := s.svc.GetLoan(loanID)
	if err != nil {
		logger.Error("fail to get outstanding balance",
			zap.Error(err),
		)
	}

	return outstandingResponseFrom(loan), nil
}

func (s *LoanBillingGRPCServer) IsDelinquent(ctx context.Context, req *v1.IsDelinquentRequest) (*v1.IsDelinquentResponse, error) {
	logger, ok := ctx.Value(o11y.LoggerKey{}).(*zap.Logger)
	if !ok {
		log.Println("no logger in ctx, fallback to no-op logger")
		logger = zap.NewNop()
	}

	loanID, err := typeid.Parse[model.LoanID](req.LoanId)
	if err != nil {
		logger.Error("fail to parse loan id",
			zap.String("requested_loan_id", req.LoanId),
		)
		return nil, err
	}

	isDelinquent, err := s.svc.CheckDelinquency(loanID, time.Now().UTC())
	if err != nil {
		logger.Error("fail to get delinquency status",
			zap.Error(err),
		)
	}

	return isDelinquentResponseFrom(isDelinquent), nil
}

func (s *LoanBillingGRPCServer) MakePayment(ctx context.Context, req *v1.MakePaymentRequest) (*v1.MakePaymentResponse, error) {
	logger, ok := ctx.Value(o11y.LoggerKey{}).(*zap.Logger)
	if !ok {
		log.Println("no logger in ctx, fallback to no-op logger")
		logger = zap.NewNop()
	}

	loanID, err := typeid.Parse[model.LoanID](req.LoanId)
	if err != nil {
		logger.Error("fail to parse loan id",
			zap.String("requested_loan_id", req.LoanId),
		)
		return nil, err
	}

	err = req.When.CheckValid()
	if err != nil {
		logger.Error("invalid payment time",
			zap.Error(err),
		)
		return nil, err
	}

	amount := currency.NewRupiah(int(req.Amount), int(req.Decimal))
	if req.Currency != amount.ISOCode() {
		logger.Error("mismatch currency",
			zap.String("requested_currency", req.Currency),
		)
		return nil, err
	}

	err = s.svc.RecordPayment(loanID, req.When.AsTime(), amount)
	if err != nil {
		logger.Error("fail to make payment",
			zap.Error(err),
		)
	}

	return &v1.MakePaymentResponse{}, nil
}
