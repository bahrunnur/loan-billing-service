package config

type ServiceConfig struct {
	Distribution string `env:"DISTRIBUTION" envDefault:"development" envDocs:"Binary distribution environment (valid: [development, production])"`

	BasePath    string `env:"BASE_PATH" envDefault:"/loan-billing-service" envDocs:"The base path for the REST api"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"loan-billing" envDocs:"The name of the service"`
	Port        int    `env:"PORT" envDefault:"8080" envDocs:"The port which the service will listen to"`
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"8081" envDocs:"The gRPC port which the service will listen to"`

	MissedPaymentThreshold int `env:"MISSED_PAYMENT_THRESHOLD" envDefault:"1" envDocs:"Missing Repayment Threshold to be flagged as delinquent account"`
}
