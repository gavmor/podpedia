package pipeline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Downloader", func() {
	var (
		downloader *Downloader
		dest       string
		content    string
		ts         *httptest.Server
	)

	BeforeEach(func() {
		downloader = NewDownloader()
		dest = "test_audio.mp3"
		content = "audio content"

		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, content)
		}))
	})

	AfterEach(func() {
		ts.Close()
		// Cleanup is implicit with MemMapFs in most tests, but if we used a real dest path,
		// we should be careful. Since dest is a string "test_audio.mp3",
		// and Downloader.DownloadAudio is tested with MemMapFs, we don't need os.Remove.
	})

	Describe("DownloadAudio", func() {
		It("downloads the audio file to the specified destination", func() {
			fs := afero.NewMemMapFs()
			err := downloader.DownloadAudio(context.Background(), fs, ts.URL, dest)
			Expect(err).NotTo(HaveOccurred())

			data, err := afero.ReadFile(fs, dest)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(content))
		})
	})
})
