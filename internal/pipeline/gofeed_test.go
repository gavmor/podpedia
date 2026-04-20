package pipeline_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/gavmor/podpedia/internal/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gofeed Integration", func() {
	Describe("ParseRSSWithGofeed", func() {
		var (
			ts       *httptest.Server
			xmlInput string
		)

		BeforeEach(func() {
			xmlInput = `
				<rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
					<channel>
						<title>Integrated Podcast</title>
						<description>A podcast for testing gofeed</description>
						<itunes:author>Test Author</itunes:author>
						<category>Technology</category>
						<category>Science</category>
						<item>
							<title>Integrated Episode</title>
							<description>Testing metadata extraction</description>
							<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
							<guid>integrated-guid</guid>
							<enclosure url="http://example.com/audio.mp3" type="audio/mpeg"/>
							<itunes:duration>00:45:00</itunes:duration>
							<itunes:explicit>yes</itunes:explicit>
							<dc:creator>Dublin Core Creator</dc:creator>
							<category>News</category>
						</item>
					</channel>
				</rss>
			`
		})

		JustBeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, xmlInput)
			}))
		})

		AfterEach(func() {
			ts.Close()
		})

		It("correctly parses the podcast and episode metadata", func() {
			podcast, episodes, err := ParseRSSWithGofeed(ts.URL)
			Expect(err).NotTo(HaveOccurred())

			Expect(podcast.Title).To(Equal("Integrated Podcast"))
			Expect(podcast.Author).To(Equal("Test Author"))
			Expect(podcast.Categories).To(ConsistOf("Technology", "Science"))

			Expect(episodes).To(HaveLen(1))
			ep := episodes[0]
			Expect(ep.Title).To(Equal("Integrated Episode"))
			Expect(ep.Duration).To(Equal("00:45:00"))
			Expect(ep.Explicit).To(BeTrue())
			Expect(ep.Author).To(Equal("Dublin Core Creator"))
			Expect(ep.Categories).To(ContainElement("News"))
		})
	})
})
