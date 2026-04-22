// plugins/downloader is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/downloader

go 1.26.1

require (
	github.com/extism/go-pdk v1.1.3
	github.com/gavmor/wasm-microkernel v0.5.2
)

replace github.com/gavmor/wasm-microkernel => ../../internal/wasm-microkernel

replace github.com/rozoomcool/go-ollama-sdk => ../../internal/sdk/go-ollama-sdk
