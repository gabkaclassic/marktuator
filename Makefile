BINARY_NAME = marktuator
BUILD_DIR = build
COVERAGE_FILE = coverage.out

README_FILE = README.md
MAN_FILE = marktuator.1

.PHONY: all build test man clean

all: build

build:
	@echo "==> Building binary..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/marktuator

test:
	@echo "==> Running tests with coverage..."
	@go clean -testcache
	@go test ./... -coverprofile=$(COVERAGE_FILE)
	@go tool cover -func=$(COVERAGE_FILE)

man:
	@echo "==> Generating man page from README.md..."
	@pandoc -s -t man $(README_FILE) -o $(MAN_FILE)

clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_FILE) $(MAN_FILE)
