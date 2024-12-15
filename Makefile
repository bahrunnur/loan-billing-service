# Output directory for binaries
BIN_DIR := bin

# Find all main.go files under cmd/<binary>
BINS := $(shell find cmd -type f -name main.go | sed 's|/main.go||' | sed 's|cmd/||')

# Build rule for each binary
$(BIN_DIR)/%: cmd/%/main.go
	mkdir -p $(BIN_DIR)
	go build -o $@ ./$<

# Default target: build all binaries
all: $(addprefix $(BIN_DIR)/, $(BINS))

# Clean up binaries
clean:
	rm -rf $(BIN_DIR)

.PHONY: all clean test proto

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