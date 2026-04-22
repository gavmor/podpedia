// plugins/store is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/store

go 1.26.1

require (
	github.com/gavmor/podpedia v0.2.2
	github.com/gavmor/wasm-microkernel v0.8.0
	github.com/samber/lo v1.53.0
)

require (
	github.com/extism/go-pdk v1.1.3 // indirect
	golang.org/x/text v0.36.0 // indirect
)

replace github.com/rozoomcool/go-ollama-sdk => ../../internal/sdk/go-ollama-sdk
