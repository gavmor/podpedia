package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "podpedia",
	Version: appVersion,
	Short:   "Podpedia - Podcast Encyclopedia Pipeline",
	Long: `Podpedia is an automated pipeline that ingests podcast RSS feeds, 
processes episodes, and extracts structured data using local LLMs to build a database 
of guests, companies, and ideologies.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Podpedia %s (%s) built at %s\n", appVersion, appCommit, appDate)
		fmt.Println("Use --help to see available commands.")
	},
}

func SetVersion(v, c, d string) {
	appVersion = v
	appCommit = c
	appDate = d
	rootCmd.Version = v
}
func Execute() error {
	return rootCmd.Execute()
}
