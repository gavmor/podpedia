package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/podpedia/pkg/podpedia"
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

		opts := []podpedia.Option{
			podpedia.WithLogger(logger),
			podpedia.WithOllamaURL(ollamaURL),
			podpedia.WithOllamaModel(ollamaModel),
			podpedia.WithTranscribeURL(transcribeURL),
		}

		if episodeLimit > 0 {
			opts = append(opts, podpedia.WithLimit(episodeLimit))
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
			opts = append(opts, podpedia.WithScheme(schemeBytes, id))
		}

		app, err := podpedia.NewApp(opts...)
		if err != nil {
			fmt.Printf("Error: failed to initialize application: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = app.Close() }()

		// Create a cancelable context tied to OS signals
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Use an explicit signal channel for logging to avoid false positives from stop()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// Listen for signals to log and allow forced exit
		go func() {
			select {
			case <-sigCh:
				logger.Info("signal-received-cancelling-gracefully", lager.Data{"hint": "press Ctrl+C again to force exit"})
				signal.Stop(sigCh) // stop listening for signals to restore default behavior
			case <-ctx.Done():
				// Context cancelled by other means
			}
		}()

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

			app.LoadPlugin(name, wasmBytes)
		}

		if err := app.Run(ctx, rssURL, absOutput); err != nil {
			if ctx.Err() != nil {
				logger.Info("run-cancelled-by-user")
				os.Exit(0)
			}
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
