package pipeline_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Downloader", func() {
	var (
		downloader *Downloader
	)

	BeforeEach(func() {
		downloader = NewDownloader()
	})

	Describe("FetchMetadata", func() {
		var (
			ts            *httptest.Server
			contentLength string
			acceptRanges  string
			statusCode    int
		)

		BeforeEach(func() {
			contentLength = "1024"
			acceptRanges = "bytes"
			statusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", contentLength)
				if acceptRanges != "" {
					w.Header().Set("Accept-Ranges", acceptRanges)
				}
				w.WriteHeader(statusCode)
			}))
		})

		AfterEach(func() {
			ts.Close()
		})

		It("correctly fetches the file size and range support", func() {
			size, supportsRange, err := downloader.FetchMetadata(ts.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(1024)))
			Expect(supportsRange).To(BeTrue())
		})

		Context("when the server does not support ranges", func() {
			BeforeEach(func() {
				acceptRanges = ""
			})

			It("reports that ranges are not supported", func() {
				_, supportsRange, err := downloader.FetchMetadata(ts.URL)
				Expect(err).NotTo(HaveOccurred())
				Expect(supportsRange).To(BeFalse())
			})
		})

		Context("when the content length is invalid", func() {
			BeforeEach(func() {
				contentLength = "not-a-number"
			})

			It("returns an error", func() {
				_, _, err := downloader.FetchMetadata(ts.URL)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DownloadAudio", func() {
		var (
			ts      *httptest.Server
			content string
			dest    string
		)

		BeforeEach(func() {
			content = "fake audio content"
			dest = "test_audio_bdd.mp3"
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, content)
			}))
		})

		AfterEach(func() {
			ts.Close()
			os.Remove(dest)
		})

		It("downloads the audio file to the specified destination", func() {
			err := downloader.DownloadAudio(ts.URL, dest)
			Expect(err).NotTo(HaveOccurred())

			data, err := os.ReadFile(dest)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(content))
		})
	})
})
