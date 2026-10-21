package podpedia_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gavmor/podpedia/pkg/podpedia"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

func TestPodpedia(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Podpedia Package Suite")
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

var _ = Describe("App", func() {
	It("can be initialized with functional options", func() {
		fs := afero.NewMemMapFs()
		app, err := podpedia.NewApp(
			podpedia.WithWorkspace(fs),
			podpedia.WithoutCache(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(app).NotTo(BeNil())
	})

	Describe("artifact cache", func() {
		var tmp string

		BeforeEach(func() {
			tmp = GinkgoT().TempDir()
			// Keep every spec off the real user cache directory.
			GinkgoT().Setenv("PODPEDIA_CACHE_DIR", filepath.Join(tmp, "env"))
		})

		It("opens the directory named by the option", func() {
			dir := filepath.Join(tmp, "explicit")

			_, err := podpedia.NewApp(podpedia.WithCacheDir(dir))

			Expect(err).NotTo(HaveOccurred())
			Expect(dirExists(dir)).To(BeTrue())
		})

		It("falls back to PODPEDIA_CACHE_DIR", func() {
			_, err := podpedia.NewApp()

			Expect(err).NotTo(HaveOccurred())
			Expect(dirExists(filepath.Join(tmp, "env"))).To(BeTrue())
		})

		It("lets the option override the environment", func() {
			dir := filepath.Join(tmp, "explicit")

			_, err := podpedia.NewApp(podpedia.WithCacheDir(dir))

			Expect(err).NotTo(HaveOccurred())
			Expect(dirExists(dir)).To(BeTrue())
			Expect(dirExists(filepath.Join(tmp, "env"))).To(BeFalse())
		})

		It("creates nothing when caching is switched off", func() {
			_, err := podpedia.NewApp(podpedia.WithoutCache())

			Expect(err).NotTo(HaveOccurred())
			Expect(dirExists(filepath.Join(tmp, "env"))).To(BeFalse())
		})

		It("treats an empty PODPEDIA_CACHE_DIR as off", func() {
			GinkgoT().Setenv("PODPEDIA_CACHE_DIR", "")

			_, err := podpedia.NewApp()

			Expect(err).NotTo(HaveOccurred())
			Expect(dirExists(filepath.Join(tmp, "env"))).To(BeFalse())
		})

		It("accepts a budget from the environment", func() {
			GinkgoT().Setenv("PODPEDIA_CACHE_MAX_BYTES", "1048576")

			_, err := podpedia.NewApp()

			Expect(err).NotTo(HaveOccurred())
			Expect(dirExists(filepath.Join(tmp, "env"))).To(BeTrue())
		})

		It("starts anyway when the budget is unparseable", func() {
			GinkgoT().Setenv("PODPEDIA_CACHE_MAX_BYTES", "lots")

			app, err := podpedia.NewApp()

			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
			Expect(dirExists(filepath.Join(tmp, "env"))).To(BeTrue())
		})

		It("starts anyway when the cache directory cannot be created", func() {
			blocked := filepath.Join(tmp, "a-file")
			Expect(os.WriteFile(blocked, []byte("not a directory"), 0o644)).To(Succeed())
			GinkgoT().Setenv("PODPEDIA_CACHE_DIR", filepath.Join(blocked, "cache"))

			app, err := podpedia.NewApp()

			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
		})
	})
})
