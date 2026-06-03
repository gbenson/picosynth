all: build

PHONY: build check lint run run-local test

check: test

lint:
	gofmt -w .
	go vet ./...

test: lint

build: check
	tinygo build -target=pico -o picosynth.uf2 ./cmd/picosynth

run: check
	tinygo flash -target=pico -monitor ./cmd/picosynth

run-local: check
	go run ./cmd/picosynth
