TARGET ?= pico2
GOFLAGS ?= -target=$(TARGET) -scheduler=cores
BINARY ?= picosynth_$(TARGET).uf2

BUILDER_IMAGE ?= gbenson/tinygo-dev:20260720
BUILDER_CACHE ?= $(HOME)/.cache/picosynth/$(subst :,_,$(BUILDER_IMAGE))

all: $(BINARY)

PHONY: check flash generate lint run test

check: test

lint:
	gofmt -w .
	go mod tidy

generate:
	go generate ./...

test: lint generate
	go vet ./...
	go test ./...

$(BINARY): check
	mkdir -p $(BUILDER_CACHE)
	docker run --rm -it \
		-v $(shell go env GOPATH):/go \
		-v $(BUILDER_CACHE):/home/user \
		-v $(PWD):/src \
		--workdir=/src \
		$(BUILDER_IMAGE) \
		./docker-entrypoint.sh \
	tinygo build $(GOFLAGS) -o $@ ./cmd/picosynth

flash: $(BINARY)
	go run ./cmd/devtool flash $<

run: check
	go run ./cmd/picosynth
