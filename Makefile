PLUGINS := rss extractor downloader transcriber store
DIST    := dist/plugins
GO      := go
WASM_FLAGS := -buildmode=c-shared

.PHONY: all plugins clean test ci $(PLUGINS)

all: plugins

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

## Remove compiled plugins
clean:
	rm -rf $(DIST)

## Full CI: build all plugins then run tests
ci: plugins test
