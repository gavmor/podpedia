// plugins/rss is an isolated Go module so the host project is not polluted by
// WASM-only dependencies (or wasip1 build constraints).

module github.com/gavmor/podpedia/plugins/rss

go 1.26.1

require (
	github.com/gavmor/wasm-microkernel v0.5.2
	github.com/mmcdole/gofeed v1.3.0
)

require (
	github.com/PuerkitoBio/goquery v1.8.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/bytecodealliance/wasm-tools-go v0.3.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mmcdole/goxpp v1.1.1-0.20240225020742-a0c311522b23 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)

replace github.com/rozoomcool/go-ollama-sdk => ../../internal/sdk/go-ollama-sdk
