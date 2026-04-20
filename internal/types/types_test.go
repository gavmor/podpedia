package types_test

import (
	"encoding/xml"

	. "github.com/gavmor/podpedia/internal/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {
	Describe("Episode", func() {
		Context("when unmarshaling from XML", func() {
			var (
				xmlInput []byte
				ep       Episode
			)

			BeforeEach(func() {
				xmlInput = []byte(`
					<item>
						<title>Episode 1</title>
						<description>The first episode</description>
						<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
						<guid>e1-guid</guid>
					</item>
				`)
			})

			JustBeforeEach(func() {
				err := xml.Unmarshal(xmlInput, &ep)
				Expect(err).NotTo(HaveOccurred())
			})

			It("correctly parses the basic fields", func() {
				Expect(ep.Title).To(Equal("Episode 1"))
				Expect(ep.Description).To(Equal("The first episode"))
				Expect(ep.ID).To(Equal("e1-guid"))
				Expect(ep.PubDate).To(Equal("Mon, 01 Jan 2024 00:00:00 +0000"))
			})

			It("initializes secondary fields as empty", func() {
				// These are not in the XML item tag directly or marked as ignored
				Expect(ep.AudioURL).To(BeEmpty())
				Expect(ep.Author).To(BeEmpty())
				Expect(ep.Duration).To(BeEmpty())
				Expect(ep.Explicit).To(BeFalse())
			})
		})
	})
})
