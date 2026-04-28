package pipeline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gofeed RSS Parsing", func() {
	var (
		ts       *httptest.Server
		xmlInput string
	)

	BeforeEach(func() {
		xmlInput = `
			<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
				<channel>
					<title>Test Podcast</title>
					<description>A test podcast description</description>
					<itunes:author>Test Author</itunes:author>
					<item>
						<title>Test Episode 1</title>
						<description>Description for episode 1</description>
						<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
						<guid>ep1</guid>
						<enclosure url="http://example.com/ep1.mp3" length="1234" type="audio/mpeg"/>
						<itunes:duration>00:30:00</itunes:duration>
					</item>
				</channel>
			</rss>
		`
	})

	JustBeforeEach(func() {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, xmlInput)
		}))
	})

	AfterEach(func() {
		ts.Close()
	})

	It("correctly parses a podcast feed", func() {
		podcast, episodes, err := ParseRSSWithGofeed(context.Background(), ts.URL)
		Expect(err).NotTo(HaveOccurred())

		Expect(podcast.Title).To(Equal("Test Podcast"))
		Expect(podcast.Author).To(Equal("Test Author"))

		Expect(episodes).To(HaveLen(1))
		Expect(episodes[0].Title).To(Equal("Test Episode 1"))
		Expect(episodes[0].AudioURL).To(Equal("http://example.com/ep1.mp3"))
		Expect(episodes[0].Duration).To(Equal("00:30:00"))
	})

	Context("when the feed is invalid", func() {
		BeforeEach(func() {
			xmlInput = "not-xml"
		})

		It("returns an error", func() {
			_, _, err := ParseRSSWithGofeed(context.Background(), ts.URL)
			Expect(err).To(HaveOccurred())
		})
	})
})
