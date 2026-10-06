TARGET ?= pico2
GOFLAGS ?= -target=$(TARGET) -scheduler=cores

BUILDER_IMAGE ?= gbenson/tinygo-dev:20260720
BUILDER_CACHE ?= $(HOME)/.cache/picosynth/$(subst :,_,$(BUILDER_IMAGE))

all: build

PHONY: build check generate lint run run-local test

check: test

lint:
	gofmt -w .
	go mod tidy

generate:
	go generate ./...

test: lint generate
	go vet ./...
	go test ./...

build: check
	mkdir -p $(BUILDER_CACHE)
	docker run --rm -it \
		-v $(shell go env GOPATH):/go \
		-v $(BUILDER_CACHE):/home/user \
		-v $(PWD):/src \
		--workdir=/src \
		$(BUILDER_IMAGE) \
		./docker-entrypoint.sh \
	tinygo build $(GOFLAGS) -o picosynth_$(TARGET).uf2 ./cmd/picosynth

flash: check
	tinygo flash $(GOFLAGS) -monitor ./cmd/picosynth

run: check
	go run ./cmd/picosynth
