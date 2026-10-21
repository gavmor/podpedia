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
		fs     afero.Fs
		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		logger = lagertest.NewTestLogger("pipeline-afero-test")
	})

	build := func(t Transcriber) *Pipeline {
		return NewPipeline(
			logger,
			t,
			&mockExtractor{},
			&countingDownloader{fs: fs},
			&writingStore{fs: fs},
		).WithWorkspace(fs).WithScheme([]byte(`{}`), "entry")
	}

	It("writes this run's output into the provided workspace", func() {
		ep := types.Episode{ID: "test-ep-1", AudioURL: "https://example.com/1.mp3"}
		p := build(&countingTranscriber{transcript: "fresh transcript"})

		Expect(p.ProcessEpisode(context.Background(), &ep, "test-output-afero")).To(Succeed())

		content, err := afero.ReadFile(fs, "test-output-afero/test_ep_1_raw.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("fresh transcript"))
	})

	// The workspace is a scratch directory the caller owns: it may arrive
	// empty on every run and be deleted afterwards, so what it happens to
	// contain is not evidence that the work was already done. Reuse lives in
	// the engine's own store instead (see WithCache), keyed by enclosure URL.
	It("does not treat workspace contents as a cache", func() {
		transcriptPath := "test-output-afero/test_ep_1_raw.txt"
		Expect(fs.MkdirAll("test-output-afero", 0o755)).To(Succeed())
		Expect(afero.WriteFile(fs, transcriptPath, []byte("stale transcript"), 0o644)).To(Succeed())

		transcriber := &countingTranscriber{transcript: "fresh transcript"}
		ep := types.Episode{ID: "test-ep-1", AudioURL: "https://example.com/1.mp3"}

		Expect(build(transcriber).ProcessEpisode(context.Background(), &ep, "test-output-afero")).To(Succeed())

		Expect(transcriber.calls).To(Equal(1), "a stale file must not stand in for the ASR stage")
		Expect(ep.Transcript).To(Equal("fresh transcript"))

		content, err := afero.ReadFile(fs, transcriptPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("fresh transcript"))
	})

	It("surfaces a transcription failure", func() {
		ep := types.Episode{ID: "test-ep-1", AudioURL: "https://example.com/1.mp3"}

		err := build(&errorTranscriber{err: fmt.Errorf("asr unavailable")}).
			ProcessEpisode(context.Background(), &ep, "test-output-afero")

		Expect(err).To(MatchError(ContainSubstring("asr unavailable")))
	})
})
