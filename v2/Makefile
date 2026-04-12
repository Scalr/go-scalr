.PHONY: build generate test lint fmt install-hooks install-tools
GOLANGCI_LINT_VERSION := v2.5.0

build: ## Build the generator
	@echo "Building generator..."
	@go build -o bin/scalr-gen ./cmd/scalr-gen
	@echo "✓ Built bin/scalr-gen"

generate: ## Generate public API client from OpenAPI spec
	@go run ./cmd/scalr-gen \
		--spec=../../fatmouse/taco/openapi/openapi-public.yml \
		--package=scalr
	@go build ./scalr/...
	@echo "✓ Generated code compiles"

test: ## Run tests
	@echo "Running tests..."
	@go test -race -v $(TESTARGS) ./internal/...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run cmd/... internal/...

fmt: ## Run formatter
	@echo "Formatting code..."
	@golangci-lint fmt cmd/... internal/...

install-tools: ## Install required development tools (run once after cloning)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b "$(shell go env GOPATH)/bin" $(GOLANGCI_LINT_VERSION)
	@echo "✓ golangci-lint $(GOLANGCI_LINT_VERSION) installed"

install-hooks: ## Install git pre-commit hook (run once after cloning)
	@git config core.hooksPath "$(CURDIR)/.hooks"
	@echo "✓ Git hooks installed ($(CURDIR)/.hooks)"
