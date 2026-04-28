package pipeline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Pipeline", func() {
	var (
		pipeline *Pipeline
		logger   *lagertest.TestLogger
		store    *mockStore
		fs       afero.Fs
		outDir   string
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		logger = lagertest.NewTestLogger("pipeline-test")
		store = &mockStore{
			fs:              fs,
			savedStructured: make(map[string][]byte),
		}
		pipeline = NewPipeline(
			logger,
			&mockTranscriber{},
			&mockExtractor{},
			&mockDownloader{fs: fs},
			store,
		).WithWorkspace(fs)
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
			err := pipeline.Run(context.Background(), ts.URL, outDir)
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

			err := pipeline.Run(context.Background(), ts.URL, outDir)
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
				err := pipeline.Run(context.Background(), ts.URL, outDir)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the feed URL is unreachable", func() {
			It("returns an error", func() {
				// Use a closed server to guarantee connection failure
				badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				badURL := badServer.URL
				badServer.Close()

				err := pipeline.Run(context.Background(), badURL, outDir)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
