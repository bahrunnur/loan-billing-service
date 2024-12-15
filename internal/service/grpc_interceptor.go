package service

import (
	"context"

	"github.com/bahrunnur/loan-billing-service/pkg/o11y"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func unaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		logger.Info("Incoming gRPC request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		)

		ctx = o11y.SetLogger(ctx, logger)
		// Call the handler
		resp, err := handler(ctx, req)

		// Log the response or error
		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.String("code", status.Code(err).String()))
		} else {
			logger.Info("gRPC request succeeded",
				zap.String("method", info.FullMethod),
				zap.Any("response", resp))
		}

		return resp, err
	}
}

func streamLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		logger.Info("Incoming gRPC streaming request",
			zap.String("method", info.FullMethod),
			zap.Bool("is_client_stream", info.IsClientStream),
			zap.Bool("is_server_stream", info.IsServerStream),
		)

		err := handler(srv, ss)

		// Log the outcome
		if err != nil {
			logger.Error("gRPC streaming request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.String("code", status.Code(err).String()))
		} else {
			logger.Info("gRPC streaming request succeeded",
				zap.String("method", info.FullMethod))
		}

		return err
	}
}
