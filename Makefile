SHELL = bash

NOMAD_PLUGIN_DIR ?= /tmp/nomad-plugins

.PHONY: clean
clean:
	@echo "==> Cleanup previous build"
	rm -f $(NOMAD_PLUGIN_DIR)/nomad-device-nvidia

.PHONY: copywrite
copywrite:
	@echo "==> Checking copywrite headers"
	copywrite --config .copywrite.hcl headers --spdx "MPL-2.0"

.PHONY: compile
compile: clean
	@echo "==> Compile nvidia driver plugin ..."
	mkdir -p $(NOMAD_PLUGIN_DIR)
	go build -race -trimpath -o $(NOMAD_PLUGIN_DIR)/nomad-device-nvidia cmd/main.go

.PHONY: test
test:
	@echo "==> Running tests ..."
	go test -v -race ./...

.PHONY: hack
hack: compile
hack:
	@echo "==> Run dev Nomad with nomad plugin"
	sudo nomad agent -dev -plugin-dir=$(NOMAD_PLUGIN_DIR)

# CRT release compilation
pkg/%/nomad-device-nvidia: GO_OUT ?= $@
pkg/%/nomad-device-nvidia:
	@echo "==> RELEASE BUILD of $@ ..."
	GOOS=linux GOARCH=$(lastword $(subst _, ,$*)) \
	go build -trimpath -o $(GO_OUT) cmd/main.go

# CRT release packaging (zip only)
.PRECIOUS: pkg/%/nomad-device-nvidia
pkg/%.zip: pkg/%/nomad-device-nvidia
	@echo "==> RELEASE PACKAGING of $@ ..."
	@cp LICENSE $(dir $<)LICENSE.txt
	zip -j $@ $(dir $<)*

# CRT version generation
.PHONY: version
version:
	@$(CURDIR)/version/generate.sh version/version.go version/version.go
