package main

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transcriber Plugin Logic", func() {
	Describe("parseASRResponse", func() {
		It("extracts the transcript from valid JSON", func() {
			raw := `{"transcript": "hello world"}`
			Expect(parseASRResponse(raw)).To(Equal("hello world"))
		})

		It("returns the raw string if JSON parsing fails", func() {
			Expect(parseASRResponse("just a string")).To(Equal("just a string"))
		})
	})

	Describe("buildTranscribeBody", func() {
		It("returns valid JSON with the audio_url", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3")
			var decoded map[string]string
			err := json.Unmarshal([]byte(res), &decoded)
			Expect(err).NotTo(HaveOccurred())
			Expect(decoded["audio_url"]).To(Equal("http://example.com/audio.mp3"))
		})
	})

	Describe("formatTranscriptResult", func() {
		It("wraps the transcript in a JSON object", func() {
			res := formatTranscriptResult("some text")
			Expect(res).To(MatchJSON(`{"transcript": "some text"}`))
		})
	})
})
