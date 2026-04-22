// plugins/downloader is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/downloader

go 1.26.1

require (
	github.com/extism/go-pdk v1.1.3
	github.com/gavmor/wasm-microkernel v0.5.2
)

require github.com/bytecodealliance/wasm-tools-go v0.3.2 // indirect
