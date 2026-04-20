package pipeline_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

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

func (m *mockExtractor) ExtractEntities(ep types.Episode) (types.EncyclopediaEntry, error) {
	return types.EncyclopediaEntry{EpisodeID: ep.ID}, nil
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
}

func (m *mockStore) SaveRawData(outputDir string, ep types.Episode) error {
	m.rawSaved = true
	return nil
}

func (m *mockStore) SaveStructuredData(outputDir string, entry types.EncyclopediaEntry) error {
	m.structuredSaved = true
	return nil
}

var _ = Describe("Pipeline", func() {
	var (
		pipeline *Pipeline
		logger   *lagertest.TestLogger
		store    *mockStore
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("pipeline-test")
		store = &mockStore{}
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
				fmt.Fprintln(w, `
					<rss version="2.0">
						<channel>
							<title>Test Podcast</title>
							<item>
								<title>Test Episode</title>
								<guid>test-ep</guid>
								<enclosure url="http://example.com/audio.mp3" type="audio/mpeg"/>
							</item>
						</channel>
					</rss>
				`)
			}))
		})

		AfterEach(func() {
			ts.Close()
			os.RemoveAll(outDir)
		})

		It("successfully runs the pipeline for a feed", func() {
			err := pipeline.Run(ts.URL, outDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(store.rawSaved).To(BeTrue())
			Expect(store.structuredSaved).To(BeTrue())
		})

		Context("when the feed is invalid", func() {
			It("returns an error", func() {
				err := pipeline.Run("cache:not-a-valid-url", outDir)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
