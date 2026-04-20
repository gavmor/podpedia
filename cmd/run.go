package cmd

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/podpedia/internal/llm"
	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/storage"
	"github.com/gavmor/podpedia/internal/transcription"
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

		logger := lager.NewLogger("podpedia")
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

		p := pipeline.NewPipeline(
			logger,
			transcription.NewTranscriber(),
			llm.NewExtractor(),
			pipeline.NewDownloader(),
			storage.NewStore(),
		)

		if err := p.Run(rssURL, outputDir); err != nil {
			logger.Error("pipeline-failed", err)
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().StringVarP(&rssURL, "url", "u", "", "URL of the podcast RSS feed")
	runCmd.Flags().StringVarP(&outputDir, "output", "o", "output", "Directory to save processed data")
	if err := runCmd.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(runCmd)
}
