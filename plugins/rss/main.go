//go:build wasip1

// Plugin: rss
//
// Receives raw RSS XML from the host and returns a structured list of podcast
// episodes as JSON. Pure computation — no host capabilities needed.
//
// TODO: wire `parseRSS` (in logic.go) to the host via the wasm-microkernel
// guest SDK once `github.com/gavmor/wasm-microkernel` v0.6.0 is published.
// The expected shape is roughly:
//
//	import "github.com/gavmor/wasm-microkernel/guest"
//	func main() { guest.Register(Execute) }
//	func Execute(reqJSON string) (string, error) { ... parseRSS ... }
package main

func main() {}
