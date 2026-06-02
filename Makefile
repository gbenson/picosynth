all: build

PHONY: build check lint run test

check: test

lint:
	gofmt -w .

test: lint

run: check
	tinygo flash -target=pico -monitor ./cmd/picosynth
