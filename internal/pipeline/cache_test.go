package pipeline_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/podpedia/pkg/naming"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

// fakeCache is an in-memory ArtifactCache that records what it was asked for.
type fakeCache struct {
	mu      sync.Mutex
	entries map[string][]byte
	keys    []string // every key Put, in order
	putErr  error
}

func newFakeCache() *fakeCache {
	return &fakeCache{entries: make(map[string][]byte)}
}

func (c *fakeCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.entries[key]
	return v, ok
}

func (c *fakeCache) Put(key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.putErr != nil {
		return c.putErr
	}
	c.keys = append(c.keys, key)
	c.entries[key] = data
	return nil
}

// countingTranscriber reports how often the ASR stage actually ran.
type countingTranscriber struct {
	calls      int
	transcript string
}

func (t *countingTranscriber) Transcribe(ctx context.Context, audioURL, prompt string) (string, error) {
	t.calls++
	if t.transcript == "" {
		return "transcript of " + audioURL, nil
	}
	return t.transcript, nil
}

// countingExtractor reports how often the LLM extraction stage actually ran.
type countingExtractor struct {
	calls int
}

func (e *countingExtractor) ExtractEntities(ctx context.Context, ep *types.Episode, scheme []byte) ([]byte, error) {
	e.calls++
	return []byte(fmt.Sprintf(`{"episode_id":%q,"scheme":%q}`, ep.ID, string(scheme))), nil
}

// countingDownloader reports how often audio was actually fetched.
type countingDownloader struct {
	calls int
	fs    afero.Fs
}

func (d *countingDownloader) DownloadAudio(ctx context.Context, fs afero.Fs, url, dest string) error {
	d.calls++
	if fs == nil {
		fs = d.fs
	}
	if err := fs.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return afero.WriteFile(fs, dest, []byte("audio"), 0o644)
}

// writingStore mirrors what the real store plugin does to the workspace, so a
// spec can assert what a harvesting caller would find there.
type writingStore struct {
	fs afero.Fs
}

func (s *writingStore) SaveRawData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode) error {
	if err := fs.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	return afero.WriteFile(fs, filepath.Join(outputDir, naming.RawName(ep.ID)), []byte(ep.Transcript), 0o644)
}

func (s *writingStore) SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry []byte, schemeID string) error {
	if err := fs.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	return afero.WriteFile(fs, filepath.Join(outputDir, naming.StructuredName(ep.ID, schemeID)), entry, 0o644)
}

var _ = Describe("Pipeline artifact cache", func() {
	const audioURL = "https://cdn.example.com/show/ep-42.mp3"

	var (
		logger      *lagertest.TestLogger
		cache       *fakeCache
		transcriber *countingTranscriber
		extractor   *countingExtractor
		downloader  *countingDownloader
		episode     types.Episode
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("pipeline-cache-test")
		cache = newFakeCache()
		transcriber = &countingTranscriber{}
		extractor = &countingExtractor{}
		downloader = &countingDownloader{}
		episode = types.Episode{
			ID:       "guid-42",
			Title:    "Episode 42",
			AudioURL: audioURL,
		}
	})

	// run processes ep in a brand-new, empty workspace — exactly what a caller
	// that isolates every job hands over — and returns that workspace.
	run := func(ep *types.Episode, scheme []byte, schemeID string) afero.Fs {
		fs := afero.NewMemMapFs()
		downloader.fs = fs
		p := NewPipeline(logger, transcriber, extractor, downloader, &writingStore{fs: fs}).
			WithWorkspace(fs).
			WithCache(cache).
			WithScheme(scheme, schemeID)
		ExpectWithOffset(1, p.ProcessEpisode(context.Background(), ep, "out")).To(Succeed())
		return fs
	}

	defaultRun := func(ep *types.Episode) afero.Fs {
		return run(ep, []byte(`{}`), "entry")
	}

	It("runs every stage on a cold cache", func() {
		defaultRun(&episode)

		Expect(downloader.calls).To(Equal(1))
		Expect(transcriber.calls).To(Equal(1))
		Expect(extractor.calls).To(Equal(1))
	})

	It("skips download and ASR when re-ingesting into a fresh workspace", func() {
		defaultRun(&episode)

		// A second job: new empty workspace, nothing carried over but the cache.
		second := types.Episode{ID: "guid-42", Title: "Episode 42", AudioURL: audioURL}
		defaultRun(&second)

		Expect(downloader.calls).To(Equal(1), "audio should not be downloaded twice")
		Expect(transcriber.calls).To(Equal(1), "ASR should not run twice")
		Expect(extractor.calls).To(Equal(1), "extraction should not run twice")
		Expect(second.Transcript).To(Equal("transcript of " + audioURL))
	})

	It("still fills the workspace on a full cache hit", func() {
		defaultRun(&episode)

		second := types.Episode{ID: "guid-42", Title: "Episode 42", AudioURL: audioURL}
		fs := defaultRun(&second)

		// A caller harvests the workspace it handed over; a recalled episode
		// has to look exactly like a freshly computed one.
		raw, err := afero.ReadFile(fs, "out/guid_42_raw.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(Equal("transcript of " + audioURL))

		entry, err := afero.ReadFile(fs, "out/guid_42_entry.json")
		Expect(err).NotTo(HaveOccurred())
		Expect(entry).To(MatchJSON(`{"episode_id":"guid-42","scheme":"{}"}`))
	})

	It("keys on the enclosure URL, not the title or GUID", func() {
		defaultRun(&episode)

		// The publisher re-titles the episode and regenerates its <guid>; the
		// audio it points at is unchanged, so neither stage should re-run.
		renamed := types.Episode{
			ID:       "guid-42-v2",
			Title:    "Episode 42 (Remastered)",
			AudioURL: audioURL,
		}
		defaultRun(&renamed)

		Expect(transcriber.calls).To(Equal(1))
		Expect(extractor.calls).To(Equal(1))
		Expect(renamed.Transcript).To(Equal("transcript of " + audioURL))
	})

	It("misses for a different enclosure URL", func() {
		defaultRun(&episode)

		other := types.Episode{ID: "guid-43", AudioURL: "https://cdn.example.com/show/ep-43.mp3"}
		defaultRun(&other)

		Expect(transcriber.calls).To(Equal(2))
		Expect(extractor.calls).To(Equal(2))
	})

	It("re-extracts under a different scheme but reuses the transcript", func() {
		run(&episode, []byte(`{"a":1}`), "alpha")

		other := types.Episode{ID: "guid-42", AudioURL: audioURL}
		run(&other, []byte(`{"b":2}`), "beta")

		Expect(transcriber.calls).To(Equal(1), "the transcript does not depend on the scheme")
		Expect(extractor.calls).To(Equal(2), "a different scheme is a different artifact")
	})

	It("uses hashed, filename-safe keys", func() {
		defaultRun(&episode)

		Expect(cache.keys).To(HaveLen(2))
		Expect(cache.keys).To(ContainElement(MatchRegexp(`^transcript-[0-9a-f]{64}$`)))
		Expect(cache.keys).To(ContainElement(MatchRegexp(`^entry-[0-9a-f]{64}$`)))
		for _, k := range cache.keys {
			Expect(k).NotTo(ContainSubstring("guid-42"))
			Expect(k).NotTo(ContainSubstring("example.com"))
		}
	})

	It("does not cache an episode with no enclosure URL", func() {
		ep := types.Episode{ID: "guid-no-audio"}
		defaultRun(&ep)

		Expect(cache.keys).To(BeEmpty())
	})

	It("completes the run when the cache cannot be written", func() {
		cache.putErr = fmt.Errorf("disk full")

		fs := defaultRun(&episode)

		Expect(episode.Transcript).To(Equal("transcript of " + audioURL))
		exists, err := afero.Exists(fs, "out/guid_42_entry.json")
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeTrue())
	})

	It("runs every stage when no cache is configured", func() {
		fs := afero.NewMemMapFs()
		downloader.fs = fs
		p := NewPipeline(logger, transcriber, extractor, downloader, &writingStore{fs: fs}).
			WithWorkspace(fs).
			WithScheme([]byte(`{}`), "entry")

		Expect(p.ProcessEpisode(context.Background(), &episode, "out")).To(Succeed())
		second := types.Episode{ID: "guid-42", AudioURL: audioURL}
		Expect(p.ProcessEpisode(context.Background(), &second, "out")).To(Succeed())

		Expect(transcriber.calls).To(Equal(2))
		Expect(extractor.calls).To(Equal(2))
	})
})
