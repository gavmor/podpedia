package pipeline_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gavmor/podpedia/internal/kernel"
	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ollama Integration", func() {
	var (
		k         *kernel.PodpediaKernel
		outputDir string
	)

	BeforeEach(func() {
		// Skip unless PODPEDIA_INTEGRATION is explicitly set
		if os.Getenv("PODPEDIA_INTEGRATION") != "true" {
			Skip("Skipping Ollama integration test. Set PODPEDIA_INTEGRATION=true to run.")
		}

		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
		model := os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "qwen2.5:0.5b"
		}

		k = kernel.NewPodpediaKernel(ollamaURL, model, "")

		outputDir = "test_integration_output"
		_ = os.MkdirAll(outputDir, 0755)
	})

	AfterEach(func() {
		_ = os.RemoveAll(outputDir)
		if k != nil {
			_ = k.Close()
		}
	})

	It("successfully extracts entities from the Go Time fixture using Ollama", func() {
		scheme := []byte(`{
			"guest_name": "string",
			"company": "string",
			"topics": "array of strings"
		}`)

		ep := types.Episode{
			ID:    "integration-test-ep",
			Title: "Go Time Integration",
			Transcript: `[00:00:00] Mat Ryer: Hello Alice Smith from Acme Corp.
[00:00:10] Alice Smith: Hi Mat, thanks for having me to talk about GopherDB.`,
		}

		// Save the transcript to the output dir to simulate processEpisode behavior
		transcriptPath := filepath.Join(outputDir, Slug(ep.ID)+"_raw.txt")
		err := os.WriteFile(transcriptPath, []byte(ep.Transcript), 0644)
		Expect(err).NotTo(HaveOccurred())

		By("calling the extractor directly")
		ext := kernel.NewExtractor(k)
		res, err := ext.ExtractEntities(context.Background(), &ep, scheme)
		Expect(err).NotTo(HaveOccurred())

		var result map[string]interface{}
		err = json.Unmarshal(res, &result)
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("\n--- Ollama Integration Result ---\n%s\n--------------------------------\n", string(res))

		Expect(result["guest_name"]).To(ContainSubstring("Alice"))
		Expect(result["company"]).To(ContainSubstring("Acme"))
	})
})
