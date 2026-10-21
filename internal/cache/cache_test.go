package cache_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gavmor/podpedia/internal/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// age backdates an entry so eviction treats it as least recently used.
func age(dir, key string, d time.Duration) {
	t := time.Now().Add(-d)
	ExpectWithOffset(1, os.Chtimes(filepath.Join(dir, key), t, t)).To(Succeed())
}

func sizeOf(dir string) int64 {
	entries, err := os.ReadDir(dir)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	var total int64
	for _, e := range entries {
		info, err := e.Info()
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		total += info.Size()
	}
	return total
}

var _ = Describe("Store", func() {
	var (
		dir   string
		store *cache.Store
	)

	BeforeEach(func() {
		dir = filepath.Join(GinkgoT().TempDir(), "artifacts")
		var err error
		store, err = cache.NewStore(dir, 1024)
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates its directory", func() {
		info, err := os.Stat(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.IsDir()).To(BeTrue())
	})

	It("requires a directory", func() {
		_, err := cache.NewStore("  ", 1024)
		Expect(err).To(HaveOccurred())
	})

	It("defaults the budget when none is given", func() {
		s, err := cache.NewStore(filepath.Join(GinkgoT().TempDir(), "d"), 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.MaxBytes()).To(Equal(cache.DefaultMaxBytes))
	})

	It("round-trips a value", func() {
		Expect(store.Put("transcript-abc", []byte("hello"))).To(Succeed())

		got, ok := store.Get("transcript-abc")
		Expect(ok).To(BeTrue())
		Expect(string(got)).To(Equal("hello"))
	})

	It("reports a miss for an unknown key", func() {
		_, ok := store.Get("transcript-nope")
		Expect(ok).To(BeFalse())
	})

	It("replaces an existing value", func() {
		Expect(store.Put("k", []byte("first"))).To(Succeed())
		Expect(store.Put("k", []byte("second"))).To(Succeed())

		got, ok := store.Get("k")
		Expect(ok).To(BeTrue())
		Expect(string(got)).To(Equal("second"))
	})

	It("survives being reopened", func() {
		Expect(store.Put("k", []byte("durable"))).To(Succeed())

		reopened, err := cache.NewStore(dir, 1024)
		Expect(err).NotTo(HaveOccurred())

		got, ok := reopened.Get("k")
		Expect(ok).To(BeTrue())
		Expect(string(got)).To(Equal("durable"))
	})

	It("leaves no partial files behind", func() {
		Expect(store.Put("k", []byte("value"))).To(Succeed())

		entries, err := os.ReadDir(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Name()).To(Equal("k"))
	})

	Describe("key validation", func() {
		DescribeTable("rejects a key that cannot safely name a file",
			func(key string) {
				Expect(store.Put(key, []byte("x"))).To(MatchError(cache.ErrKeyInvalid))
				_, ok := store.Get(key)
				Expect(ok).To(BeFalse())
			},
			Entry("empty", ""),
			Entry("parent directory", ".."),
			Entry("path separator", "a/b"),
			Entry("traversal", "../escape"),
			Entry("hidden", ".hidden"),
			Entry("temp-file namespace", ".tmp-x"),
			Entry("over-long", strings.Repeat("a", 201)),
		)

		It("accepts the hex keys the pipeline derives", func() {
			key := "transcript-" + strings.Repeat("0f", 32)
			Expect(store.Put(key, []byte("x"))).To(Succeed())
			_, ok := store.Get(key)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("eviction", func() {
		It("refuses a value larger than the whole budget", func() {
			err := store.Put("huge", make([]byte, 2048))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("budget"))

			_, ok := store.Get("huge")
			Expect(ok).To(BeFalse())
		})

		It("still enforces the budget when it refuses an oversize value", func() {
			Expect(store.Put("a", make([]byte, 400))).To(Succeed())
			Expect(store.Put("b", make([]byte, 400))).To(Succeed())

			// A shrunken budget: nothing new fits, but the existing entries
			// must not stay over it forever.
			shrunk, err := cache.NewStore(dir, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(shrunk.Put("c", make([]byte, 400))).To(HaveOccurred())

			Expect(sizeOf(dir)).To(BeNumerically("<=", 100))
		})

		It("keeps the directory within its budget", func() {
			for _, key := range []string{"a", "b", "c", "d", "e", "f"} {
				Expect(store.Put(key, make([]byte, 400))).To(Succeed())
			}
			Expect(sizeOf(dir)).To(BeNumerically("<=", 1024))
		})

		It("discards the least recently used entry first", func() {
			Expect(store.Put("old", make([]byte, 400))).To(Succeed())
			Expect(store.Put("new", make([]byte, 400))).To(Succeed())
			age(dir, "old", 2*time.Hour)

			// Pushes the directory to 1200 bytes against a 1024 budget, so
			// exactly one entry has to go.
			Expect(store.Put("newest", make([]byte, 400))).To(Succeed())

			_, ok := store.Get("old")
			Expect(ok).To(BeFalse())
			_, ok = store.Get("new")
			Expect(ok).To(BeTrue())
			_, ok = store.Get("newest")
			Expect(ok).To(BeTrue())
		})

		It("spares an entry a read has just refreshed", func() {
			Expect(store.Put("stale", make([]byte, 400))).To(Succeed())
			Expect(store.Put("revived", make([]byte, 400))).To(Succeed())
			age(dir, "stale", 2*time.Hour)
			age(dir, "revived", 3*time.Hour)

			_, ok := store.Get("revived") // refreshes recency
			Expect(ok).To(BeTrue())

			Expect(store.Put("newest", make([]byte, 400))).To(Succeed())

			_, ok = store.Get("stale")
			Expect(ok).To(BeFalse())
			_, ok = store.Get("revived")
			Expect(ok).To(BeTrue())
		})

		It("reclaims an orphaned partial write once it is old enough", func() {
			orphan := filepath.Join(dir, ".tmp-orphan")
			Expect(os.WriteFile(orphan, make([]byte, 100), 0o644)).To(Succeed())
			old := time.Now().Add(-2 * time.Hour)
			Expect(os.Chtimes(orphan, old, old)).To(Succeed())

			Expect(store.Put("k", []byte("x"))).To(Succeed())

			_, err := os.Stat(orphan)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("leaves a fresh partial write alone", func() {
			inflight := filepath.Join(dir, ".tmp-inflight")
			Expect(os.WriteFile(inflight, make([]byte, 100), 0o644)).To(Succeed())

			Expect(store.Put("k", []byte("x"))).To(Succeed())

			_, err := os.Stat(inflight)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("tolerates concurrent writers", func() {
		done := make(chan error, 8)
		for i := 0; i < 8; i++ {
			key := string(rune('a' + i))
			go func() {
				done <- store.Put(key, make([]byte, 100))
			}()
		}
		for i := 0; i < 8; i++ {
			Expect(<-done).To(Succeed())
		}
		Expect(sizeOf(dir)).To(BeNumerically("<=", 1024))
	})
})
