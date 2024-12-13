SERVICE_BINARY_NAME=loanbilling

build:
	go build -o $(SERVICE_BINARY_NAME)

test:
	go test ./...