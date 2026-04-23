package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Downloader Plugin Logic", func() {
	Describe("validateDownloadRequest", func() {
		DescribeTable("valid requests",
			func(url, dest string) {
				Expect(validateDownloadRequest(url, dest)).To(Succeed())
			},
			Entry("standard http", "http://example.com/ep.mp3", "/tmp/ep.mp3"),
			Entry("secure https", "https://cdn.example.com/audio.mp3", "output/audio.mp3"),
		)

		Context("with missing fields", func() {
			It("errors if URL is missing", func() {
				err := validateDownloadRequest("", "/tmp/out.mp3")
				Expect(err).To(MatchError("url required"))
			})

			It("errors if destination is missing", func() {
				err := validateDownloadRequest("http://example.com/ep.mp3", "")
				Expect(err).To(MatchError("dest required"))
			})

			It("prioritizes URL error over destination error", func() {
				err := validateDownloadRequest("", "")
				Expect(err).To(MatchError("url required"))
			})
		})

		DescribeTable("invalid URL schemes",
			func(url string) {
				err := validateDownloadRequest(url, "/tmp/out.mp3")
				Expect(err).To(MatchError("url must be http(s)"))
			},
			Entry("ftp", "ftp://example.com/ep.mp3"),
			Entry("local file", "file:///local/path.mp3"),
			Entry("no scheme", "just-a-filename.mp3"),
			Entry("absolute path", "/absolute/path.mp3"),
		)
	})
})
