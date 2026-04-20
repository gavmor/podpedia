package cmd

import (
	"fmt"
	"os"

	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/spf13/cobra"
)

var (
	rssURL    string
	outputDir string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the podcast encyclopedia pipeline",
	Long:  `Starts the full pipeline: Ingestion -> Transcription -> Entity Extraction -> Storage`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if rssURL == "" {
			fmt.Println("Error: RSS feed URL is required. Use --url or -u")
			return
		}
		if err := pipeline.Run(rssURL, outputDir); err != nil {
			fmt.Printf("Pipeline error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().StringVarP(&rssURL, "url", "u", "", "URL of the podcast RSS feed")
	runCmd.Flags().StringVarP(&outputDir, "output", "o", "output", "Directory to save processed data")
	runCmd.MarkFlagRequired("url")
	rootCmd.AddCommand(runCmd)
}
