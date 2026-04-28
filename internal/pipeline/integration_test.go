package pipeline_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/gavmor/podpedia/internal/kernel"
	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ollama Integration", func() {
	var (
		k         *kernel.PodpediaKernel
		p         *pipeline.Pipeline
		logger    *lagertest.TestLogger
		outputDir string
	)

	BeforeEach(func() {
		// Skip unless explicitly enabled with ENABLE_OLLAMA_INTEGRATION_TESTS
		if os.Getenv("ENABLE_OLLAMA_INTEGRATION_TESTS") == "" {
			Skip("Ollama integration tests disabled by default. Set ENABLE_OLLAMA_INTEGRATION_TESTS=1 to enable.")
		}

		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}

		logger = lagertest.NewTestLogger("integration-test")
		outputDir = "test_integration_output"

		// Use a more capable model for integration tests
		model := "qwen3.5:latest"

		k = kernel.NewPodpediaKernel(ollamaURL, model, "")

		// Load real plugins from dist
		plugins := []string{"rss", "downloader", "transcriber", "extractor", "store"}
		for _, name := range plugins {
			path := fmt.Sprintf("../../dist/plugins/%s.wasm", name)
			wasmBytes, err := os.ReadFile(path)
			if err != nil {
				Skip(fmt.Sprintf("WASM plugins not found at %s. Run 'make plugins' first.", path))
			}
			k.Load(name, wasmBytes)
		}

		p = pipeline.NewPipeline(
			logger,
			kernel.NewTranscriber(k),
			kernel.NewExtractor(k),
			kernel.NewDownloader(k),
			kernel.NewStore(k),
		)
	})

	AfterEach(func() {
		if k != nil {
			_ = k.Close()
		}
		_ = os.RemoveAll(outputDir)
	})

	It("successfully extracts entities from the Go Time fixture using Ollama", func() {
		ep := types.Episode{
			ID:         "integration-test-ep",
			Title:      "Systems Programming in Go",
			Transcript: GoTimeTranscriptFixture,
		}

		// Save the transcript fixture to skip transcription stage and provide input
		_ = os.MkdirAll(outputDir, 0755)
		transcriptPath := filepath.Join(outputDir, "integration_test_ep_raw.txt")
		err := os.WriteFile(transcriptPath, []byte(GoTimeTranscriptFixture), 0644)
		Expect(err).NotTo(HaveOccurred())

		// We run processEpisode directly to test the extraction logic
		// Use a simple custom scheme
		scheme := []byte(`{"guest_name": "", "company": ""}`)
		p.WithScheme(scheme, "integration")

		// Capture stdout to see plugin logs if any
		_ = p.Run("dummy", outputDir) // Run will fail because dummy isn't a URL, but we just want to test if k.Call works
		// Actually, let's just call the extractor directly via the kernel adapter

		ext := kernel.NewExtractor(k)
		res, err := ext.ExtractEntities(ep, scheme)
		Expect(err).NotTo(HaveOccurred())

		var result map[string]any
		err = json.Unmarshal(res, &result)
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("\n--- Ollama Integration Result ---\n%s\n--------------------------------\n", string(res))

		Expect(result["guest_name"]).To(ContainSubstring("Alice"))
		Expect(result["company"]).To(ContainSubstring("Acme"))
	})
})
