package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bahrunnur/loan-billing-service/internal/config"
	"github.com/bahrunnur/loan-billing-service/internal/service"
	"github.com/bahrunnur/loan-billing-service/pkg/o11y"
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := env.ParseAs[config.ServiceConfig]()
	if err != nil {
		log.Fatal(err.Error())
	}

	var logger *zap.Logger
	switch cfg.Distribution {
	case "production":
		l, err := zap.NewProduction()
		if err != nil {
			log.Fatal(err.Error())
		}
		logger = l
	default:
		l, err := zap.NewDevelopment()
		if err != nil {
			log.Fatal(err.Error())
		}
		logger = l
	}

	ctx = context.WithValue(ctx, o11y.LoggerKey{}, logger)

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownSignal
		logger.Info("got shutdown signal")
		cancel()
	}()

	service.Run(ctx, cfg)
}
