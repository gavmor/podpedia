package main

import (
	"github.com/gavmor/podpedia/cmd"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
