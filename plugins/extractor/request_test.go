package main

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// hostRequestJSON is a byte-for-byte sample of what
// internal/kernel.(*extractor).ExtractEntities marshals and hands to this
// plugin. Keep it a literal rather than re-marshalling types.Episode here:
// the point of the test is to pin the *wire* keys, so a literal fails if the
// struct this module compiles against stops matching them.
const hostRequestJSON = `{"episode":{` +
	`"id":"fx-ep-04",` +
	`"title":"The One About Pins",` +
	`"audio_url":"https://example.test/fx-ep-04.mp3",` +
	`"pub_date":"Mon, 20 Oct 2026 09:00:00 +0000",` +
	`"description":"A short description.",` +
	`"author":"Fixture Author",` +
	`"duration":"00:12:34",` +
	`"explicit":true,` +
	`"categories":["Technology"],` +
	`"transcript":"The transcript body."` +
	`},"scheme":null}`

var _ = Describe("extractRequest", func() {
	Describe("decoding the host's episode envelope", func() {
		var req extractRequest

		BeforeEach(func() {
			req = extractRequest{}
			Expect(json.Unmarshal([]byte(hostRequestJSON), &req)).To(Succeed())
		})

		// These three are what the plugin actually reads. They bind by
		// case-insensitive name even against a tagless struct, so they were
		// never affected by the stale v0.2.2 pin — assert them so a future
		// rename is caught.
		It("binds the fields the extractor reads", func() {
			Expect(req.Episode.Title).To(Equal("The One About Pins"))
			Expect(req.Episode.Transcript).To(Equal("The transcript body."))
			Expect(req.Episode.Description).To(Equal("A short description."))
		})

		// These two bind only through the json tags types.Episode grew after
		// v0.2.2. Go's default matching is case-insensitive but does not
		// cross underscores, so they silently decoded as "" for as long as
		// this module compiled against the published v0.2.2 zip. This is the
		// assertion that fails if the `replace ... => ../..` is ever dropped.
		It("binds the underscored keys that require the engine's json tags", func() {
			Expect(req.Episode.AudioURL).To(Equal("https://example.test/fx-ep-04.mp3"),
				"audio_url bound to nothing: plugins/extractor is compiling types.Episode from a published podpedia zip, not from ../..")
			Expect(req.Episode.PubDate).To(Equal("Mon, 20 Oct 2026 09:00:00 +0000"),
				"pub_date bound to nothing: plugins/extractor is compiling types.Episode from a published podpedia zip, not from ../..")
		})

		It("binds the remaining metadata", func() {
			Expect(req.Episode.ID).To(Equal("fx-ep-04"))
			Expect(req.Episode.Author).To(Equal("Fixture Author"))
			Expect(req.Episode.Duration).To(Equal("00:12:34"))
			Expect(req.Episode.Explicit).To(BeTrue())
			Expect(req.Episode.Categories).To(Equal([]string{"Technology"}))
		})
	})
})
