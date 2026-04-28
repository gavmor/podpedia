package pipeline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Pipeline Progress Hooks", func() {
	var (
		fs       afero.Fs
		pipeline *Pipeline
		logger   *lagertest.TestLogger
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		logger = lagertest.NewTestLogger("pipeline-progress-test")
	})

	It("triggers OnEpisodeComplete for each episode", func() {
		pipeline = NewPipeline(
			logger,
			&mockTranscriber{},
			&mockExtractor{},
			&mockDownloader{fs: fs},
			&mockStore{fs: fs, savedStructured: make(map[string][]byte)},
		).WithWorkspace(fs)

		var mu sync.Mutex
		completedEpisodes := make(map[string]error)

		pipeline.OnEpisodeComplete(func(ep types.Episode, err error) {
			mu.Lock()
			completedEpisodes[ep.ID] = err
			mu.Unlock()
		})

		xmlInput := `
			<rss version="2.0">
				<channel>
					<title>Test</title>
					<item><title>ep1</title><guid>ep1</guid><enclosure url="url1" type="audio/mpeg"/></item>
					<item><title>ep2</title><guid>ep2</guid><enclosure url="url2" type="audio/mpeg"/></item>
				</channel>
			</rss>
		`
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, xmlInput)
		}))
		defer ts.Close()

		err := pipeline.Run(context.Background(), ts.URL, "output")
		Expect(err).NotTo(HaveOccurred())

		Expect(completedEpisodes).To(HaveLen(2))
		Expect(completedEpisodes).To(HaveKey("ep1"))
		Expect(completedEpisodes).To(HaveKey("ep2"))
		Expect(completedEpisodes["ep1"]).To(BeNil())
		Expect(completedEpisodes["ep2"]).To(BeNil())
	})
})
