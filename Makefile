PLUGINS := rss extractor downloader transcriber store
DIST    := dist/plugins
GO      := go
WASM_FLAGS := -buildmode=c-shared

.PHONY: all plugins clean test test-plugins ci $(PLUGINS)

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

## Run unit tests for all plugin logic (no WASM build needed)
test-plugins:
	@for p in $(PLUGINS); do \
		echo "[test] plugin: $$p"; \
		cd plugins/$$p && $(GO) test -v ./... && cd ../..; \
	done

## Remove compiled plugins
clean:
	rm -rf $(DIST)

## Full CI: build all plugins then run all tests
ci: plugins test test-plugins
