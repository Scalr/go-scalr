.PHONY: build generate test lint

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
	@go test -v $(TESTARGS) ./internal/...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run internal/... || true
