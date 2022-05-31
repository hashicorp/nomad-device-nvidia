SHELL = bash
default: help

GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)

GO_LDFLAGS := "-X github.com/hashicorp/nomad-autoscaler/version.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)"

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""

.PHONY: clean
clean: ## Cleanup previous build
	@echo "==> Cleanup previous build"
	rm -f ./bin/nomad-device-nvidia

pkg/%/nomad-device-nvidia: GO_OUT ?= $@
pkg/%/nomad-device-nvidia: ## Build nomad-device-nvidia for GOOS_GOARCH, e.g. pkg/linux_amd64/nomad-device-nvidia
	@echo "==> Building $@ with tags $(GO_TAGS)..."
	@GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build -trimpath -ldflags $(GO_LDFLAGS) -tags "$(GO_TAGS)" -o $(GO_OUT) cmd/main.go

.PRECIOUS: pkg/%/nomad-device-nvidia
pkg/%.zip: pkg/%/nomad-device-nvidia ## Build and zip nomad-device-nvidia for GOOS_GOARCH, e.g. pkg/linux_amd64.zip
	@echo "==> Packaging for $@..."
	zip -j $@ $(dir $<)*

.PHONY: dev
dev: clean bin/nomad-device-nvidia ## Build the nomad-device-nvidia plugin

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
ifneq (,$(wildcard version/version_ent.go))
	@$(CURDIR)/scripts/version.sh version/version.go version/version_ent.go
else
	@$(CURDIR)/scripts/version.sh version/version.go version/version.go
endif
