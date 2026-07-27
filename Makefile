# Wisp CLI — developer entrypoints
# Prefer these targets over ad-hoc go commands so CI and local stay aligned.

GO        ?= go
GOWORK    ?= off
export GOWORK
GOPROXY   ?= https://proxy.golang.org,direct
export GOPROXY
GOSUMDB   ?= sum.golang.org
export GOSUMDB

PKG       := ./cmd/... ./internal/... ./pkg/...
BIN_DIR   := bin
BIN       := $(BIN_DIR)/wisp
COVER_OUT := coverage.out

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mUsage\033[0m\n  make \033[36m<target>\033[0m\n\n\033[1mTargets\033[0m\n"} \
		/^[a-zA-Z0-9_.-]+:.*?##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: submodules
submodules: ## Init / update git submodules (sdk)
	git submodule update --init --recursive

.PHONY: tidy
tidy: ## go mod tidy (public modules via proxy)
	$(GO) mod tidy

.PHONY: verify
verify: ## Verify module checksums against the sumdb
	$(GO) mod verify

.PHONY: fmt
fmt: ## Format with gofmt + goimports (if installed)
	$(GO) fmt $(PKG)
	@command -v goimports >/dev/null 2>&1 && goimports -w cmd internal pkg || true

.PHONY: vet
vet: ## go vet
	$(GO) vet $(PKG)

.PHONY: lint
lint: ## golangci-lint (v2 config in .golangci.yml)
	golangci-lint run --timeout=5m $(PKG)

.PHONY: test
test: ## Unit tests with race + shuffle
	$(GO) test -race -count=1 -shuffle=on $(PKG)

.PHONY: cover
cover: ## Tests with coverage report
	$(GO) test -race -count=1 -covermode=atomic -coverprofile=$(COVER_OUT) $(PKG)
	$(GO) tool cover -func=$(COVER_OUT) | tail -1

VERSION ?= dev
LDFLAGS ?= -s -w -X github.com/wisp-trading/wisp/cmd.Version=$(VERSION)

.PHONY: build
build: ## Build CLI → bin/wisp (VERSION=vX.Y.Z optional)
	mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN) .

.PHONY: smoke
smoke: build ## Green path: CLI + reference standalone binary
	cd examples/reference-standalone && $(GO) mod download && $(GO) build -trimpath -o reference-standalone .
	@echo ""
	@echo "OK: $(BIN) and examples/reference-standalone/reference-standalone"
	@echo "Run (needs keys in ~/.wisp/connectors.yml via: wisp → Settings):"
	@echo "  ./examples/reference-standalone/reference-standalone \\"
	@echo "    --config ./examples/reference-standalone"

.PHONY: ci
ci: verify vet lint test smoke ## Local approximation of GitHub CI

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) $(COVER_OUT) examples/reference-standalone/reference-standalone
