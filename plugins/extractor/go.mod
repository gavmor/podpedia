// plugins/extractor is an isolated Go module so the host project is not
// polluted by WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/extractor

go 1.26.1

require (
	github.com/gavmor/podpedia v0.2.2
	github.com/gavmor/wasm-microkernel v0.12.1
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.1
	github.com/rozoomcool/go-ollama-sdk v0.0.0-20250620220025-710cf9a2c767
	github.com/samber/lo v1.53.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/extism/go-pdk v1.1.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260115054156-294ebfa9ad83 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
)

replace github.com/rozoomcool/go-ollama-sdk => ../../internal/sdk/go-ollama-sdk
