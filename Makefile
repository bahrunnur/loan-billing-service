SERVICE_BINARY_NAME=loanbilling

build:
	go build -o $(SERVICE_BINARY_NAME)

test:
	go test ./...

proto-lint: # @HELP lints the protobuf files
proto-lint:
	echo "Running buf lint"
	docker run --rm -v $(CURDIR)/proto:/proto --workdir /proto bufbuild/buf lint

proto-gen: # @HELP generates go code from proto files
proto-gen:
	echo "Running buf generate"
	docker run --rm -v $(CURDIR)/proto:/proto --workdir /proto bufbuild/buf generate

protobuf: # @HELP lints and generates go code from proto files
protobuf: proto-lint proto-gen