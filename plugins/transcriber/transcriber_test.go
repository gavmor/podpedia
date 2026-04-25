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
		It("always includes audio_url", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3", "")
			var decoded map[string]string
			err := json.Unmarshal([]byte(res), &decoded)
			Expect(err).NotTo(HaveOccurred())
			Expect(decoded["audio_url"]).To(Equal("http://example.com/audio.mp3"))
		})

		It("omits initial_prompt when prompt is empty", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3", "")
			var decoded map[string]json.RawMessage
			Expect(json.Unmarshal([]byte(res), &decoded)).To(Succeed())
			Expect(decoded).NotTo(HaveKey("initial_prompt"))
		})

		It("forwards prompt as initial_prompt", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3", "Ada Lovelace on computing")
			var decoded map[string]string
			Expect(json.Unmarshal([]byte(res), &decoded)).To(Succeed())
			Expect(decoded["initial_prompt"]).To(Equal("Ada Lovelace on computing"))
		})
	})

	Describe("formatTranscriptResult", func() {
		It("wraps the transcript in a JSON object", func() {
			res := formatTranscriptResult("some text")
			Expect(res).To(MatchJSON(`{"transcript": "some text"}`))
		})
	})
})
