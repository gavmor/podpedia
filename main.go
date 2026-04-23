package main

import (
	"embed"
	"os"

	"github.com/gavmor/podpedia/cmd"
)

//go:embed dist/plugins/*.wasm
var plugins embed.FS

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.DefaultPlugins = plugins
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
