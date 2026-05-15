FIXTURES_DIR := $(CURDIR)/tests/fixtures/ethereum-tests
EEST_DIR := $(CURDIR)/tests/fixtures/execution-spec-tests
EEST_FIXTURES := $(EEST_DIR)/fixtures
EEST_VERSION := v5.4.0
ETHEREUM_TESTS_VERSION := v17.1
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || printf "%s/bin/golangci-lint" "$$(go env GOPATH)")

.PHONY: all test test-unit test-spec lint download-lint ethereum-tests-fixtures eest-fixtures

# Run all tests, including EEST fixtures.
all: test

# Run all fixture tests (GeneralStateTests, BlockchainTests, TransactionTests, EEST)
test: test-unit test-spec

# Run unit tests (no fixtures needed)
test-unit:
	go test ./internal/... -count=1

# Run Go lint checks.
lint:
	@test -x "$(GOLANGCI_LINT)" || { \
		echo "golangci-lint is not installed. Run: make download-lint"; \
		exit 1; \
	}
	$(GOLANGCI_LINT) run ./...

# Install golangci-lint into GOBIN or GOPATH/bin.
download-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all ethereum spec fixture tests
test-spec: ethereum-tests-fixtures eest-fixtures
	GEVM_TESTS_DIR=$(FIXTURES_DIR)/GeneralStateTests \
	GEVM_BLOCKCHAIN_TESTS_DIR=$(FIXTURES_DIR)/BlockchainTests \
	GEVM_TRANSACTION_TESTS_DIR=$(FIXTURES_DIR)/TransactionTests \
	GEVM_EEST_DIR=$(EEST_FIXTURES)/state_tests \
	go test ./tests/spec/... -count=1 -timeout=30m -failfast

# Download ethereum/tests fixtures with a sparse checkout.
ethereum-tests-fixtures:
	@if [ ! -d "$(FIXTURES_DIR)/GeneralStateTests" ] || [ ! -d "$(FIXTURES_DIR)/BlockchainTests" ] || [ ! -d "$(FIXTURES_DIR)/TransactionTests" ]; then \
		echo "Downloading ethereum/tests fixtures..."; \
		rm -rf "$(FIXTURES_DIR)"; \
		mkdir -p "$(dir $(FIXTURES_DIR))"; \
		git clone --depth=1 --filter=blob:none --sparse --branch "$(ETHEREUM_TESTS_VERSION)" https://github.com/ethereum/tests.git "$(FIXTURES_DIR)"; \
		cd "$(FIXTURES_DIR)" && git sparse-checkout set GeneralStateTests BlockchainTests TransactionTests; \
	fi

# Download EEST fixtures from GitHub release
eest-fixtures:
	@if [ ! -d "$(EEST_FIXTURES)/state_tests" ]; then \
		echo "Downloading EEST fixtures $(EEST_VERSION)..."; \
		mkdir -p "$(EEST_DIR)"; \
		curl -sL https://github.com/ethereum/execution-spec-tests/releases/download/$(EEST_VERSION)/fixtures_stable.tar.gz | \
		tar xz -C $(EEST_DIR); \
		echo "EEST fixtures extracted to $(EEST_FIXTURES)"; \
	fi
