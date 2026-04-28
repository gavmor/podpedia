BINARY_NAME=podpedia
MAIN_PATH=.
PLUGINS := rss extractor downloader transcriber store
DIST    := dist/plugins
GO      := go
WASM_FLAGS := -buildmode=c-shared

.PHONY: all build run plugins plugins-build clean test test-plugins ci lint $(PLUGINS)

all: plugins build

## Build host binary into bin/
build: plugins-build
	@echo "[build] host: $(BINARY_NAME)"
	@mkdir -p bin
	$(GO) build -o bin/$(BINARY_NAME) $(MAIN_PATH)

## Run host binary
run: build
	./bin/$(BINARY_NAME) run

## Build all WASM plugins with go mod tidy
plugins:
	@for p in $(PLUGINS); do \
		echo "[tidy] plugin: $$p"; \
		cd plugins/$$p && $(GO) mod tidy && cd ../..; \
	done

## Internal target: Build WASM plugins without tidying
plugins-build:
	@mkdir -p $(DIST)
	@for p in $(PLUGINS); do \
		echo "[build] plugin: $$p"; \
		cd plugins/$$p && \
			GOOS=wasip1 GOARCH=wasm $(GO) build $(WASM_FLAGS) -o ../../$(DIST)/$$p.wasm .; \
		cd ../..; \
	done

## Run host unit tests (no WASM build needed)
test:
	$(GO) test ./internal/...

## Run unit tests for all plugin logic (no WASM build needed)
test-plugins:
	@for p in $(PLUGINS); do \
		echo "[test] plugin: $$p"; \
		cd plugins/$$p && $(GO) test -v ./... && cd ../..; \
	done

## Run linter (requires plugins for go:embed)
lint: plugins-build
	golangci-lint run

## Remove compiled artifacts
clean:
	rm -rf dist bin

## Full CI: build everything then run all tests
ci: build test test-plugins lint
