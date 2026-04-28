package pipeline_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Pipeline Error Aggregation", func() {
	var (
		fs       afero.Fs
		pipeline *Pipeline
		logger   *lagertest.TestLogger
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		logger = lagertest.NewTestLogger("pipeline-error-test")
	})

	It("aggregates multiple errors using errors.Join", func() {
		pipeline = NewPipeline(
			logger,
			&errorTranscriber{err: ErrTranscriptionTimeout},
			&mockExtractor{},
			&mockDownloader{fs: fs},
			&mockStore{fs: fs, savedStructured: make(map[string][]byte)},
		).WithWorkspace(fs)

		// Create a local server to serve a feed with multiple episodes
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
		Expect(err).To(HaveOccurred())

		// It should be a joined error containing ErrTranscriptionTimeout
		joined, ok := err.(interface{ Unwrap() []error })
		Expect(ok).To(BeTrue(), "expected errors.Join result")
		causes := joined.Unwrap()
		Expect(len(causes)).To(BeNumerically(">=", 1))
		for _, cause := range causes {
			Expect(errors.Is(cause, ErrTranscriptionTimeout)).To(BeTrue())
		}
	})
})
