all: build

PHONY: build check lint run test

check: test

lint:
	gofmt -w .

test: lint

build: check
	tinygo build -target=pico -o picosynth.uf2 ./cmd/picosynth

run: check
	tinygo flash -target=pico -monitor ./cmd/picosynth
