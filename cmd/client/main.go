package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bahrunnur/loan-billing-service/internal/config"
	"github.com/bahrunnur/loan-billing-service/internal/model"
	v1 "github.com/bahrunnur/loan-billing-service/proto/gen/loanbilling/v1"
	"github.com/caarlos0/env/v11"
	"go.jetify.com/typeid"
	"google.golang.org/grpc"
)

// Client wraps the gRPC client connection and any specific service clients.
type Client struct {
	conn    *grpc.ClientConn
	Service v1.LoanBillingServiceClient
}

// NewClient creates a new gRPC client for the given service.
func NewClient(ctx context.Context, target string, opts ...grpc.DialOption) (*Client, error) {
	// use default options if none are provided.
	if len(opts) == 0 {
		// for production, use `WithTransportCredentials` instead of `WithInsecure`. and remove WithBlock
		opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		Service: v1.NewLoanBillingServiceClient(conn),
	}, nil
}

// Close closes the gRPC client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := env.ParseAs[config.ServiceConfig]()
	if err != nil {
		log.Fatal(err.Error())
	}

	target := fmt.Sprintf("%s:%d", "localhost", cfg.GRPCPort)
	client, err := NewClient(ctx, target)
	if err != nil {
		log.Fatalf("fail to create client: %v", err)
	}
	defer client.Close()

	loanID, err := typeid.New[model.LoanID]()
	if err != nil {
		log.Fatalf("fail to create loan type id: %v", err)
	}

	req := &v1.IsDelinquentRequest{LoanId: loanID.String()}
	resp, err := client.Service.IsDelinquent(ctx, req)
	if err != nil {
		log.Fatalf("fail to request IsDelinquent: %v", err)
	}

	log.Printf("loanID: %s, delinquent? %t", loanID, resp.IsDelinquent)
}
