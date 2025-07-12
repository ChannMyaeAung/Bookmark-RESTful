.DEFAULT_GOAL := build

.PHONY: fmt vet build
## -gcflags="-m"

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build