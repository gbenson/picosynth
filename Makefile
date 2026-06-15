GOFLAGS ?= -target=pico -scheduler=cores

all: build

PHONY: build check generate lint run run-local test

check: test

lint:
	gofmt -w .
	go vet ./...

generate:
	go generate ./...

test: lint generate
	go test ./...

build: check
	tinygo build $(GOFLAGS) -o picosynth.uf2 ./cmd/picosynth

flash: check
	tinygo flash $(GOFLAGS) -monitor ./cmd/picosynth

run: check
	go run ./cmd/picosynth
