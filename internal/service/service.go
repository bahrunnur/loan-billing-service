package service

import (
	"context"
	"fmt"
	"net"

	"github.com/bahrunnur/loan-billing-service/internal/adapters/grpchandler"
	"github.com/bahrunnur/loan-billing-service/internal/adapters/memorystorage"
	"github.com/bahrunnur/loan-billing-service/internal/config"
	"github.com/bahrunnur/loan-billing-service/internal/loan"
	"github.com/bahrunnur/loan-billing-service/pkg/o11y"
	v1 "github.com/bahrunnur/loan-billing-service/proto/gen/loanbilling/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, serviceConfig config.ServiceConfig) {
	logger := o11y.LoggerFromContext(ctx)

	// service
	storage := memorystorage.NewLoanMemoryStorage()
	loanService := loan.NewLoanService(storage)
	grpcHandler := grpchandler.NewLoanBillingGRPCServer(loanService)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", serviceConfig.GRPCPort))
	if err != nil {
		logger.Error("fail to listen",
			zap.Error(err),
		)
		return
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryLoggingInterceptor(logger)),
		grpc.StreamInterceptor(streamLoggingInterceptor(logger)),
	}

	s := grpc.NewServer(opts...)
	v1.RegisterLoanBillingServiceServer(s, grpcHandler)

	go func() {
		<-ctx.Done()
		logger.Info("stopping gRPC server")
		if s != nil {
			s.GracefulStop()
		}
	}()

	logger.Info(fmt.Sprintf("grpc server listening at %s", listen.Addr()))
	if err := s.Serve(listen); err != nil {
		logger.Fatal("fail to serve grpc endpoint",
			zap.Error(err),
		)
	}
}
