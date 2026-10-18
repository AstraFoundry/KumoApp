SHELL := /bin/bash
MYGO := ./node_modules/.bin/mygo

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show available commands.
	@awk 'BEGIN {FS = ":.*##"; printf "Kumo development commands:\n\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

node_modules: package.json bun.lock
	bun install
	@touch node_modules

.PHONY: dev
dev: node_modules ## Run the app with live reload.
	$(MYGO) dev

.PHONY: generate
generate: node_modules ## Regenerate the TypeScript client and message catalogs.
	$(MYGO) generate

.PHONY: build
build: node_modules ## Build the app for this platform into build/.
	$(MYGO) build

.PHONY: cli
cli: ## Build only the kumo command into build/kumo.
	go build -o build/kumo ./cmd/kumo

.PHONY: test
test: ## Run Go tests (set KUMO_TEST_MIHOMO=<mihomo> for the lifecycle simulations).
	go vet ./...
	go test ./...

.PHONY: lint
lint: node_modules ## Typecheck the frontend and check FSD layering and messages.
	bun run typecheck
	bun run lint:fsd
	bun run lint:i18n

.PHONY: e2e
e2e: build ## Drive the built Linux app through its lifecycle (MIHOMO=<mihomo>).
	scripts/e2e/lifecycle.sh $(MIHOMO)

.PHONY: check
check: generate test lint ## Everything CI runs before building.
