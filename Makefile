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
	tinygo build -target=pico -o picosynth.uf2 ./cmd/picosynth

run: check
	tinygo flash -target=pico -monitor ./cmd/picosynth

run-local: check
	go run ./cmd/picosynth
