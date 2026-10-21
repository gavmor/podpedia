package pipeline_test

import (
	"context"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/gavmor/podpedia/internal/cache"
	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

// These specs wire the pipeline to the real on-disk store rather than a fake,
// so the two halves of the feature are exercised together: the pipeline's key
// derivation has to produce keys the store will accept, and the store's
// contents have to survive a process that keeps nothing else between runs.
var _ = Describe("Pipeline with the on-disk artifact cache", func() {
	const audioURL = "https://cdn.example.com/show/ep-7.mp3"

	var (
		logger      *lagertest.TestLogger
		cacheRoot   string
		transcriber *countingTranscriber
		extractor   *countingExtractor
		downloader  *countingDownloader
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("pipeline-disk-cache-test")
		cacheRoot = filepath.Join(GinkgoT().TempDir(), "artifacts")
		transcriber = &countingTranscriber{}
		extractor = &countingExtractor{}
		downloader = &countingDownloader{}
	})

	// runJob mimics a caller that isolates every job: a fresh workspace, a
	// freshly opened cache, nothing else carried over.
	runJob := func(ep *types.Episode, maxBytes int64) afero.Fs {
		store, err := cache.NewStore(cacheRoot, maxBytes)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		fs := afero.NewMemMapFs()
		downloader.fs = fs
		p := NewPipeline(logger, transcriber, extractor, downloader, &writingStore{fs: fs}).
			WithWorkspace(fs).
			WithCache(store).
			WithScheme([]byte(`{}`), "entry")

		ExpectWithOffset(1, p.ProcessEpisode(context.Background(), ep, "out")).To(Succeed())
		return fs
	}

	It("reuses work across jobs that share nothing but the cache directory", func() {
		first := types.Episode{ID: "guid-7", AudioURL: audioURL}
		runJob(&first, 1<<20)

		second := types.Episode{ID: "guid-7", AudioURL: audioURL}
		fs := runJob(&second, 1<<20)

		Expect(downloader.calls).To(Equal(1))
		Expect(transcriber.calls).To(Equal(1))
		Expect(extractor.calls).To(Equal(1))

		raw, err := afero.ReadFile(fs, "out/guid_7_raw.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(Equal("transcript of " + audioURL))
	})

	It("writes nothing but its own entries into the cache directory", func() {
		ep := types.Episode{ID: "guid-7", AudioURL: audioURL}
		runJob(&ep, 1<<20)

		entries, err := os.ReadDir(cacheRoot)
		Expect(err).NotTo(HaveOccurred())

		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		Expect(names).To(HaveLen(2))
		Expect(names).To(ContainElement(MatchRegexp(`^transcript-[0-9a-f]{64}$`)))
		Expect(names).To(ContainElement(MatchRegexp(`^entry-[0-9a-f]{64}$`)))
	})

	It("recomputes after the budget evicts the entry", func() {
		first := types.Episode{ID: "guid-7", AudioURL: audioURL}
		runJob(&first, 1<<20)

		// A budget smaller than a single entry: the next write is refused and
		// the existing entries are evicted, so the following job starts cold.
		filler := types.Episode{ID: "guid-8", AudioURL: "https://cdn.example.com/show/ep-8.mp3"}
		runJob(&filler, 1)

		again := types.Episode{ID: "guid-7", AudioURL: audioURL}
		runJob(&again, 1<<20)

		Expect(transcriber.calls).To(Equal(3))
		Expect(again.Transcript).To(Equal("transcript of " + audioURL))
	})
})
