package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/podpedia/internal/kernel"
	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/spf13/cobra"
)

var DefaultPlugins embed.FS

var (
	rssURL        string
	outputDir     string
	pluginDir     string
	ollamaURL     string
	ollamaModel   string
	transcribeURL string
	episodeLimit  int
	outputScheme  string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the podcast encyclopedia pipeline",
	Long: `Starts the full pipeline via WASM plugins:
  RSS parse → Audio download → Transcription → Entity extraction → Storage

Each stage is a sandboxed WASM plugin. Uses embedded defaults unless --plugins is provided.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if rssURL == "" {
			fmt.Println("Error: RSS feed URL is required. Use --url or -u")
			return
		}

		absOutput, err := filepath.Abs(outputDir)
		if err != nil {
			fmt.Printf("Error: failed to get absolute path for %q: %v\n", outputDir, err)
			os.Exit(1)
		}
		if err := os.MkdirAll(absOutput, 0755); err != nil {
			fmt.Printf("Error: failed to create or access output directory %q: %v\n", absOutput, err)
			os.Exit(1)
		}

		logger := lager.NewLogger("podpedia")
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

		k := kernel.NewPodpediaKernel(ollamaURL, ollamaModel, transcribeURL)
		defer func() { _ = k.Close() }()

		plugins := []string{"rss", "downloader", "transcriber", "extractor", "store"}

		for _, name := range plugins {
			var wasmBytes []byte
			var err error

			// 1. Try to load from provided --plugins directory first
			if cmd.Flags().Changed("plugins") {
				path := fmt.Sprintf("%s/%s.wasm", pluginDir, name)
				wasmBytes, err = os.ReadFile(path)
				if err != nil {
					logger.Info("plugin-load-disk-failed-falling-back", lager.Data{"plugin": name, "path": path, "error": err.Error()})
				}
			}

			// 2. Fall back to embedded plugins if not loaded from disk
			if len(wasmBytes) == 0 {
				embedPath := fmt.Sprintf("dist/plugins/%s.wasm", name)
				wasmBytes, err = DefaultPlugins.ReadFile(embedPath)
				if err != nil {
					logger.Error("plugin-load-embedded-failed", err, lager.Data{"plugin": name, "path": embedPath})
					os.Exit(1)
				}
			}

			k.Load(name, wasmBytes)
		}

		p := pipeline.NewPipeline(
			logger,
			kernel.NewTranscriber(k),
			kernel.NewExtractor(k),
			kernel.NewDownloader(k),
			kernel.NewStore(k),
		)

		if episodeLimit > 0 {
			p.WithLimit(episodeLimit)
		}

		if outputScheme != "" {
			schemeBytes, err := os.ReadFile(outputScheme)
			if err != nil {
				fmt.Printf("Error: failed to read output scheme file %q: %v\n", outputScheme, err)
				os.Exit(1)
			}
			id := filepath.Base(outputScheme)
			if ext := filepath.Ext(id); ext != "" {
				id = id[:len(id)-len(ext)]
			}
			p.WithScheme(schemeBytes, id)
		}

		if err := p.Run(rssURL, absOutput); err != nil {
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
	runCmd.Flags().StringVar(&ollamaModel, "ollama-model", "qwen2.5:0.5b", "Ollama model for LLM inference")
	runCmd.Flags().StringVar(&transcribeURL, "transcribe-url", "", "ASR endpoint URL for transcription (Whisper.cpp, Deepgram, etc.)")
	runCmd.Flags().IntVarP(&episodeLimit, "limit", "n", 0, "Maximum number of episodes to process (0 = all)")
	runCmd.Flags().StringVar(&outputScheme, "output-scheme", "", "Path to a JSON file defining the structured output schema")
	if err := runCmd.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(runCmd)
}
