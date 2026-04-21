// plugins/downloader is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/downloader

go 1.26.1
