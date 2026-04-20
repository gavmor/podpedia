package cmd

import (
	"context"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/podpedia/internal/kernel"
	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/spf13/cobra"
)

var (
	rssURL        string
	outputDir     string
	pluginDir     string
	ollamaURL     string
	transcribeURL string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the podcast encyclopedia pipeline",
	Long: `Starts the full pipeline via WASM plugins:
  RSS parse → Audio download → Transcription → Entity extraction → Storage

Each stage is a sandboxed WASM plugin loaded from --plugins.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if rssURL == "" {
			fmt.Println("Error: RSS feed URL is required. Use --url or -u")
			return
		}

		logger := lager.NewLogger("podpedia")
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

		ctx := context.Background()

		k, err := kernel.New(ctx, logger, ollamaURL, transcribeURL)
		if err != nil {
			logger.Error("kernel-init", err)
			os.Exit(1)
		}
		defer k.Close()

		plugins := []string{"rss", "downloader", "transcriber", "extractor", "store"}
		for _, name := range plugins {
			path := fmt.Sprintf("%s/%s.wasm", pluginDir, name)
			if err := k.Load(name, path); err != nil {
				logger.Error("plugin-load", err, lager.Data{"plugin": name, "path": path})
				os.Exit(1)
			}
		}

		p := pipeline.NewPipeline(
			logger,
			kernel.NewTranscriber(k),
			kernel.NewExtractor(k),
			kernel.NewDownloader(k),
			kernel.NewStore(k),
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
	runCmd.Flags().StringVarP(&pluginDir, "plugins", "p", "dist/plugins", "Directory containing compiled .wasm plugins")
	runCmd.Flags().StringVar(&ollamaURL, "ollama", "http://localhost:11434", "Ollama base URL for LLM inference")
	runCmd.Flags().StringVar(&transcribeURL, "transcribe-url", "", "ASR endpoint URL for transcription (Whisper.cpp, Deepgram, etc.)")
	if err := runCmd.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(runCmd)
}
