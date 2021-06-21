default: build

build: bin/nomad-device-nvidia

bin/nomad-device-nvidia:
	mkdir -p bin
	go build \
		-o ./bin/nomad-device-nvidia \
		./cmd/main.go

.PHONY: test
test:
	go test -v ./...
