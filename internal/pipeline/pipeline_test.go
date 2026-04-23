package pipeline_test

import (
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
	return GoTimeTranscriptFixture, nil
}

type mockExtractor struct{}

func (m *mockExtractor) ExtractEntities(ep types.Episode, scheme []byte) ([]byte, error) {
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
	RawSaved        bool
	StructuredSaved bool
	savedStructured map[string][]byte
}

func (m *mockStore) SaveRawData(outputDir string, ep types.Episode) error {
	m.RawSaved = true
	return nil
}

func (m *mockStore) SaveStructuredData(outputDir string, ep types.Episode, entry []byte, schemeID string) error {
	m.StructuredSaved = true
	
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
		outDir   string
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
		outDir = "test_output_bdd"
	})

	AfterEach(func() {
		_ = os.RemoveAll(outDir)
	})

	Describe("Run", func() {
		var (
			ts       *httptest.Server
			xmlInput string
		)

		BeforeEach(func() {
			xmlInput = `
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
			`
		})

		JustBeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintln(w, xmlInput)
			}))
		})

		AfterEach(func() {
			ts.Close()
		})

		It("successfully executes all stages of the pipeline", func() {
			By("running the pipeline")
			err := pipeline.Run(ts.URL, outDir)
			Expect(err).NotTo(HaveOccurred())

			By("verifying the data was saved")
			Expect(store).To(SatisfyAll(
				HaveField("RawSaved", BeTrue()),
				HaveField("StructuredSaved", BeTrue()),
			))
		})

		It("correctly propagates a custom scheme and names the file correctly", func() {
			scheme := []byte(`{"sentiment": "positive"}`)
			schemeID := "sentiment"
			
			By("setting a custom scheme")
			pipeline.WithScheme(scheme, schemeID)
			
			err := pipeline.Run(ts.URL, outDir)
			Expect(err).NotTo(HaveOccurred())

			By("verifying the specific file was 'saved'")
			expectedKey := "test_ep_1_sentiment"
			Expect(store.savedStructured).To(HaveKey(expectedKey))
			
			By("verifying the content of the saved data")
			Expect(store.savedStructured[expectedKey]).To(MatchJSON(`{"episode_id": "test-ep-1", "mocked": true}`))
		})

		Context("when the feed is invalid", func() {
			BeforeEach(func() {
				xmlInput = "not-xml"
			})

			It("returns an error", func() {
				err := pipeline.Run(ts.URL, outDir)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the feed URL is unreachable", func() {
			It("returns an error", func() {
				err := pipeline.Run("http://invalid-and-unreachable-url-12345.com", outDir)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
