package pipeline_test

import (
	"context"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Pipeline with Context Cancellation", func() {
	var (
		fs       afero.Fs
		pipeline *Pipeline
		logger   *lagertest.TestLogger
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		logger = lagertest.NewTestLogger("pipeline-context-test")

		pipeline = NewPipeline(
			logger,
			&mockTranscriber{},
			&mockExtractor{},
			&mockDownloader{fs: fs},
			&mockStore{fs: fs, savedStructured: make(map[string][]byte)},
		).WithWorkspace(fs)
	})

	It("stops processing when context is cancelled", func() {
		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		err := pipeline.Run(ctx, "http://example.com/rss", "output")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("context canceled"))
	})
})
