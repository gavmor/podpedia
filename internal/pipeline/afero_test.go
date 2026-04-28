package pipeline_test

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

type errorTranscriber struct {
	err error
}

func (e *errorTranscriber) Transcribe(ctx context.Context, audioURL string, prompt string) (string, error) {
	return "", e.err
}

var _ = Describe("Pipeline with Afero Workspace", func() {
	var (
		fs       afero.Fs
		pipeline *Pipeline
		logger   *lagertest.TestLogger
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		logger = lagertest.NewTestLogger("pipeline-afero-test")

		pipeline = NewPipeline(
			logger,
			&mockTranscriber{},
			&mockExtractor{},
			&mockDownloader{fs: fs},
			&mockStore{fs: fs, savedStructured: make(map[string][]byte)},
		)
	})

	It("respects the provided workspace for transcription caching", func() {
		transcriptPath := "test-output-afero/test_ep_1_raw.txt"
		expectedTranscript := "cached transcript"

		// Pre-populate the memory filesystem
		Expect(fs.MkdirAll("test-output-afero", 0755)).To(Succeed())
		Expect(afero.WriteFile(fs, transcriptPath, []byte(expectedTranscript), 0644)).To(Succeed())

		pipeline.WithWorkspace(fs)

		ep := types.Episode{
			ID: "test-ep-1",
		}

		// We use a transcriber that returns an error to ensure we're using the cache
		eTranscriber := &errorTranscriber{err: fmt.Errorf("should not be called")}

		// Re-init pipeline with the error transcriber
		pipeline = NewPipeline(
			logger,
			eTranscriber,
			&mockExtractor{},
			&mockDownloader{fs: fs},
			&mockStore{fs: fs, savedStructured: make(map[string][]byte)},
		).WithWorkspace(fs)

		err := pipeline.ProcessEpisode(context.Background(), &ep, "test-output-afero")
		Expect(err).NotTo(HaveOccurred())
		Expect(ep.Transcript).To(Equal(expectedTranscript))
	})
})
