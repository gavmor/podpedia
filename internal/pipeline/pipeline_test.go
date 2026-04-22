package pipeline_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Mock implementations for testing
type mockTranscriber struct{}

func (m *mockTranscriber) Transcribe(audioURL string) (string, error) {
	return "mock transcript", nil
}

type mockExtractor struct{}

func (m *mockExtractor) ExtractEntities(ep types.Episode, scheme []byte) ([]byte, error) {
	// Return a mock JSON that matches the "ideology" or "entry" expectations
	return []byte(`{"episode_id": "` + ep.ID + `", "mocked": true}`), nil
}

type mockDownloader struct{}

func (m *mockDownloader) DownloadAudio(url string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.WriteFile(dest, []byte("mock audio"), 0644)
}

type mockStore struct {
	rawSaved        bool
	structuredSaved bool
	savedStructured map[string][]byte
}

func (m *mockStore) SaveRawData(outputDir string, ep types.Episode) error {
	m.rawSaved = true
	return nil
}

func (m *mockStore) SaveStructuredData(outputDir string, ep types.Episode, entry []byte, schemeID string) error {
	m.structuredSaved = true
	
	id := ep.ID
	if id == "" { id = "unknown" }
	
	key := fmt.Sprintf("%s_%s", testSlug(id), schemeID)
	m.savedStructured[key] = entry
	return nil
}

func testSlug(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
}

var _ = Describe("Pipeline", func() {
	var (
		pipeline *Pipeline
		logger   *lagertest.TestLogger
		store    *mockStore
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("pipeline-test")
		store = &mockStore{
			savedStructured: make(map[string][]byte),
		}
		pipeline = NewPipeline(
			logger,
			&mockTranscriber{},
			&mockExtractor{},
			&mockDownloader{},
			store,
		)
	})

	Describe("Run", func() {
		var (
			ts     *httptest.Server
			outDir string
		)

		BeforeEach(func() {
			outDir = "test_output_bdd"
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, `
					<rss version="2.0">
						<channel>
							<title>Test Podcast</title>
							<item>
								<title>Test Episode</title>
								<guid>test-ep-1</guid>
								<enclosure url="http://example.com/audio.mp3" type="audio/mpeg"/>
							</item>
						</channel>
					</rss>
				`)
			}))
		})

		AfterEach(func() {
			ts.Close()
			_ = os.RemoveAll(outDir)
		})

		It("successfully runs the pipeline for a feed", func() {
			err := pipeline.Run(ts.URL, outDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(store.rawSaved).To(BeTrue())
			Expect(store.structuredSaved).To(BeTrue())
		})

		It("correctly propagates a custom scheme and names the file correctly", func() {
			scheme := []byte(`{"sentiment": "positive"}`)
			schemeID := "sentiment"
			
			pipeline.WithScheme(scheme, schemeID)
			
			err := pipeline.Run(ts.URL, outDir)
			Expect(err).NotTo(HaveOccurred())

			expectedKey := "test_ep_1_sentiment"
			Expect(store.savedStructured).To(HaveKey(expectedKey))
			
			var result map[string]any
			err = json.Unmarshal(store.savedStructured[expectedKey], &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["episode_id"]).To(Equal("test-ep-1"))
		})

		Context("when the feed is invalid", func() {
			It("returns an error", func() {
				err := pipeline.Run("cache:not-a-valid-url", outDir)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
