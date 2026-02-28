IMAGE ?= backrest-volsync-operator:dev

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: generate
generate:
	@echo "No code generation configured (CRDs are tracked in config/)."

.PHONY: lint
ifeq ($(OS),Windows_NT)
GOLANGCI_LINT := $(shell where golangci-lint 2>NUL)
else
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)
endif

ifeq ($(strip $(GOLANGCI_LINT)),)
lint:
	@echo "Lint skipped (golangci-lint not installed)."
else
lint:
	golangci-lint run ./...
endif

.PHONY: test
test:
	go test ./...

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) .
