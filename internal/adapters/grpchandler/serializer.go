package grpchandler

import (
	"github.com/bahrunnur/loan-billing-service/internal/model"
	v1 "github.com/bahrunnur/loan-billing-service/proto/gen/loanbilling/v1"
)

func outstandingResponseFrom(loan model.WeeklyLoanFullInformation) *v1.GetOutstandingResponse {
	return &v1.GetOutstandingResponse{
		OutstandingBalance: int64(loan.OutstandingBalance.Rupiah()),
		Decimal:            int32(loan.OutstandingBalance.Sen()),
		Currency:           loan.OutstandingBalance.ISOCode(),
	}
}

func isDelinquentResponseFrom(isDelinquent bool) *v1.IsDelinquentResponse {
	return &v1.IsDelinquentResponse{
		IsDelinquent: isDelinquent,
	}
}
