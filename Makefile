BINARY_NAME=podpedia
MAIN_PATH=.
PLUGINS := rss extractor downloader transcriber store
DIST    := dist/plugins
GO      := go
WASM_FLAGS := -buildmode=c-shared

.PHONY: all build run plugins clean test test-plugins ci $(PLUGINS)

all: plugins build

## Build host binary into bin/
build: plugins
	@echo "[build] host: $(BINARY_NAME)"
	@mkdir -p bin
	$(GO) build -o bin/$(BINARY_NAME) $(MAIN_PATH)

## Run host binary
run: build
	./bin/$(BINARY_NAME) run

## Build all WASM plugins into dist/plugins/
plugins: $(PLUGINS)

$(PLUGINS):
	@echo "[build] plugin: $@"
	@mkdir -p $(DIST)
	cd plugins/$@ && \
		$(GO) mod tidy && \
		GOOS=wasip1 GOARCH=wasm $(GO) build $(WASM_FLAGS) -o ../../$(DIST)/$@.wasm .

## Run host unit tests (no WASM build needed)
test:
	$(GO) test ./internal/...

## Run unit tests for all plugin logic (no WASM build needed)
test-plugins:
	@for p in $(PLUGINS); do \
		echo "[test] plugin: $$p"; \
		cd plugins/$$p && $(GO) test -v ./... && cd ../..; \
	done

## Remove compiled artifacts
clean:
	rm -rf dist bin

## Full CI: build everything then run all tests
ci: build plugins test test-plugins
