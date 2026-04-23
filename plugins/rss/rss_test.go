package main

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const minimalFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Podcast</title>
    <description>A test podcast</description>
    <item>
      <title>Episode One</title>
      <guid>ep-001</guid>
      <description>  First episode.  </description>
      <enclosure url="https://example.com/ep1.mp3" type="audio/mpeg"/>
      <pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`

const multiepisodeFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
  <channel>
    <title>Multi Pod</title>
    <description>Several episodes</description>
    <item>
      <title>Ep A</title>
      <guid>ep-a</guid>
      <enclosure url="https://example.com/a.mp3" type="audio/mpeg"/>
      <itunes:duration>30:00</itunes:duration>
    </item>
    <item>
      <title>Ep B</title>
      <guid>ep-b</guid>
      <enclosure url="https://example.com/b.mp3" type="audio/mpeg"/>
    </item>
    <item>
      <title></title>
      <guid>no-title</guid>
      <enclosure url="https://example.com/c.mp3" type="audio/mpeg"/>
    </item>
    <item>
      <title>No Enclosure</title>
      <guid>no-enc</guid>
    </item>
  </channel>
</rss>`

var _ = Describe("RSS Plugin Logic", func() {
	Describe("parseRSS", func() {
		Context("with a minimal valid feed", func() {
			It("correctly parses podcast metadata", func() {
				pod, _, err := parseRSS(minimalFeed)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Title).To(Equal("Test Podcast"))
				Expect(pod.Description).To(Equal("A test podcast"))
			})

			It("correctly parses episode fields", func() {
				_, eps, err := parseRSS(minimalFeed)
				Expect(err).NotTo(HaveOccurred())
				Expect(eps).To(HaveLen(1))

				ep := eps[0]
				Expect(ep.ID).To(Equal("ep-001"))
				Expect(ep.Title).To(Equal("Episode One"))
				Expect(ep.Description).To(Equal("First episode."))
				Expect(ep.AudioURL).To(Equal("https://example.com/ep1.mp3"))
				Expect(ep.PubDate).NotTo(BeEmpty())
			})
		})

		Context("with multiple episodes and incomplete items", func() {
			It("filters out items with no title or no enclosure", func() {
				_, eps, err := parseRSS(multiepisodeFeed)
				Expect(err).NotTo(HaveOccurred())
				Expect(eps).To(HaveLen(2))

				ids := []string{eps[0].ID, eps[1].ID}
				Expect(ids).To(ContainElements("ep-a", "ep-b"))
			})

			It("parses iTunes duration if available", func() {
				_, eps, err := parseRSS(multiepisodeFeed)
				Expect(err).NotTo(HaveOccurred())

				var epA outEpisode
				for _, ep := range eps {
					if ep.ID == "ep-a" {
						epA = ep
					}
				}
				Expect(epA.Duration).To(Equal("30:00"))
			})
		})

		Context("when GUID is missing", func() {
			It("falls back to the enclosure URL as ID", func() {
				feed := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>X</title>
  <item>
    <title>No GUID Episode</title>
    <enclosure url="https://example.com/noguid.mp3" type="audio/mpeg"/>
  </item>
</channel></rss>`
				_, eps, err := parseRSS(feed)
				Expect(err).NotTo(HaveOccurred())
				Expect(eps).To(HaveLen(1))
				Expect(eps[0].ID).To(Equal("https://example.com/noguid.mp3"))
			})
		})

		Context("with an empty feed", func() {
			It("parses the podcast title but returns no episodes", func() {
				feed := `<?xml version="1.0"?><rss version="2.0"><channel><title>Empty</title></channel></rss>`
				pod, eps, err := parseRSS(feed)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Title).To(Equal("Empty"))
				Expect(eps).To(BeEmpty())
			})
		})

		It("returns an error for invalid XML", func() {
			_, _, err := parseRSS("not xml at all")
			Expect(err).To(HaveOccurred())
		})

		It("produces JSON-serializable output", func() {
			pod, eps, err := parseRSS(minimalFeed)
			Expect(err).NotTo(HaveOccurred())

			out, err := json.Marshal(map[string]any{"podcast": pod, "episodes": eps})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("Test Podcast"))
		})
	})
})
