package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "podpedia",
	Short: "Podpedia - Podcast Encyclopedia Pipeline",
	Long:  `Podpedia is an automated pipeline that ingests podcast RSS feeds, 
processes episodes, and extracts structured data using local LLMs to build a database 
of guests, companies, and ideologies.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Podpedia! Use --help to see available commands.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
