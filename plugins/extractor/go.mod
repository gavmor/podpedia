// plugins/extractor is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/extractor

go 1.26.1

require (
	github.com/gavmor/podpedia v0.2.2
	github.com/gavmor/wasm-microkernel v0.5.2
	github.com/rozoomcool/go-ollama-sdk v0.0.0-20250620220025-710cf9a2c767
	github.com/samber/lo v1.53.0
)

require (
	github.com/bytecodealliance/wasm-tools-go v0.3.2 // indirect
	golang.org/x/text v0.36.0 // indirect
)
