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
	docker run --rm -it \
		-v $(shell go env GOPATH):/go \
		-v $(PWD):/src \
		--workdir=/src \
		gbenson/tinygo-dev:20260720 \
		./docker-entrypoint.sh \
	tinygo build $(GOFLAGS) -o picosynth.uf2 ./cmd/picosynth

flash: check
	tinygo flash $(GOFLAGS) -monitor ./cmd/picosynth

run: check
	go run ./cmd/picosynth
