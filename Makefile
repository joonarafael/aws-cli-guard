SHELL := /bin/sh

GOLANGCI_LINT_MODULE := github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.14.0

.PHONY: help build build-race test vet tidy lint lint-fix lint-config file-length clean check

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?##' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

build: ## Build the wrapper binary (no cgo)
	CGO_ENABLED=0 go build -o bin/aws ./cmd/aws

build-race: ## Build the wrapper binary with the race detector
	mkdir -p bin
	CGO_ENABLED=1 go build -race -o bin/aws ./cmd/aws

test: ## Run unit tests with race detector
	CGO_ENABLED=1 go test ./... -race -count=1

vet: ## Run go vet (unix and windows build tags)
	go vet ./...
	GOOS=windows go vet ./...

tidy: ## Tidy go.mod / go.sum
	go mod tidy

lint-config: ## Validate .golangci.yml against the pinned golangci-lint schema
	./scripts/lint/golangci-config.sh

GOLANGCI_LINT := $(shell go tool -n golangci-lint)

lint: lint-config ## Run golangci-lint (check only)
	go tool golangci-lint run ./...
	GOOS=windows GOARCH=amd64 $(GOLANGCI_LINT) run ./internal/...

lint-fix: lint-config ## Run golangci-lint with --fix
	go tool golangci-lint run ./... --fix
	GOOS=windows GOARCH=amd64 $(GOLANGCI_LINT) run ./internal/... --fix

file-length: ## Fail if any .go file exceeds 300 lines
	./scripts/lint/go-file-length.sh

clean: ## Remove build artifacts
	rm -f bin/aws

check: lint file-length vet test ## Run lint, file-length, vet, and tests
