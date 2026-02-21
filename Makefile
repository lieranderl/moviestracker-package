SHELL := /bin/bash

APP_NAME ?= moviestracker
PKGS := ./...
GO_CACHE_DIR := $(CURDIR)/.cache/go-build
GO_MOD_CACHE_DIR := $(CURDIR)/.cache/go-mod
GOLANGCI_CACHE_DIR := $(CURDIR)/.cache/golangci-lint
GO_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
GOLANGCI_LINT ?= golangci-lint

.PHONY: fmt lint vet test race cover build tidy ci precommit hooks

fmt:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	gofmt -w ./cmd ./executor ./internal ./pkg

lint:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR) $(GOLANGCI_CACHE_DIR)
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	$(GO_ENV) GOLANGCI_LINT_CACHE=$(GOLANGCI_CACHE_DIR) $(GOLANGCI_LINT) run ./...

vet:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go vet $(PKGS)

test:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go test $(PKGS)

race:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go test -race $(PKGS)

cover:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go test $(PKGS) -covermode=atomic -coverprofile=coverage.out
	$(GO_ENV) go tool cover -func=coverage.out

build:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR) ./bin
	$(GO_ENV) go build -o ./bin/$(APP_NAME) ./cmd

tidy:
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go mod tidy -go=1.25.0

ci: fmt vet test race cover build

precommit:
	@command -v pre-commit >/dev/null 2>&1 || (echo "pre-commit not found. Install with: pipx install pre-commit (or brew install pre-commit)" && exit 1)
	pre-commit run --all-files

hooks:
	@command -v pre-commit >/dev/null 2>&1 || (echo "pre-commit not found. Install with: pipx install pre-commit (or brew install pre-commit)" && exit 1)
	pre-commit install
