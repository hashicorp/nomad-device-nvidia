SHELL = bash

default: help

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""

.PHONY:clean
clean: ## Cleanup previous build
	@echo "==> Cleanup previous build"
	rm -f ./bin/nomad-device-nvidia

.PHONY: build
build: clean bin/nomad-device-nvidia ## Build the nomad-device-nvidia plugin

bin/nomad-device-nvidia:
	@echo "==> Building device driver ..."
	mkdir -p bin
	go build -o bin/nomad-device-nvidia cmd/main.go

.PHONY: test
test: ## Run unit tests
	@echo "==> Running tests ..."
	go test -v -race ./...

.PHONY: version
version: ## Get the current version string
	@$(CURDIR)/scripts/version.sh version/version.go
