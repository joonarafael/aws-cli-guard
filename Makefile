SHELL := /bin/sh

GOLANGCI_LINT_MODULE := github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.14.0

# GitHub release tag is v$(RUMDL_VERSION). scripts/lint/rumdl.sh downloads it.
RUMDL_VERSION ?= 0.3.0
export RUMDL_VERSION

.PHONY: help build build-race test vet tidy lint lint-fix lint-config file-length md-lint md-fmt clean check

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?##' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

build: ## Build the wrapper and aws-guard binaries (no cgo)
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/aws ./cmd/aws
	CGO_ENABLED=0 go build -o bin/aws-guard ./cmd/aws-guard

build-race: ## Build the wrapper and aws-guard binaries with the race detector
	mkdir -p bin
	CGO_ENABLED=1 go build -race -o bin/aws ./cmd/aws
	CGO_ENABLED=1 go build -race -o bin/aws-guard ./cmd/aws-guard

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

md-lint: ## Lint Markdown with pinned rumdl (check only)
	./scripts/lint/rumdl.sh check --deny-config-warnings .

md-fmt: ## Auto-fix Markdown with pinned rumdl
	./scripts/lint/rumdl.sh fmt --deny-config-warnings .

clean: ## Remove build artifacts
	rm -f bin/aws bin/aws-guard

check: lint file-length md-lint vet test ## Run Go lint, file-length, Markdown lint, vet, and tests
