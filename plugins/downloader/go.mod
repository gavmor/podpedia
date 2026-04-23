// plugins/downloader is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/downloader

go 1.26.1

require (
	github.com/gavmor/wasm-microkernel v0.9.0
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.1
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/extism/go-pdk v1.1.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260115054156-294ebfa9ad83 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
)

replace github.com/rozoomcool/go-ollama-sdk => ../../internal/sdk/go-ollama-sdk
