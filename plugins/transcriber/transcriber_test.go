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
			res := buildTranscribeBody("http://example.com/audio.mp3", "", "")
			var decoded map[string]string
			err := json.Unmarshal([]byte(res), &decoded)
			Expect(err).NotTo(HaveOccurred())
			Expect(decoded["audio_url"]).To(Equal("http://example.com/audio.mp3"))
		})

		It("omits initial_prompt when both title and notes are empty", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3", "", "")
			var decoded map[string]json.RawMessage
			Expect(json.Unmarshal([]byte(res), &decoded)).To(Succeed())
			Expect(decoded).NotTo(HaveKey("initial_prompt"))
		})

		It("sets initial_prompt to title when only title is provided", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3", "My Show Title", "")
			var decoded map[string]string
			Expect(json.Unmarshal([]byte(res), &decoded)).To(Succeed())
			Expect(decoded["initial_prompt"]).To(Equal("Episode: My Show Title"))
		})

		It("includes both title and notes in initial_prompt", func() {
			res := buildTranscribeBody("http://example.com/audio.mp3", "Ep 42", "Guest: Ada Lovelace")
			var decoded map[string]string
			Expect(json.Unmarshal([]byte(res), &decoded)).To(Succeed())
			Expect(decoded["initial_prompt"]).To(ContainSubstring("Ep 42"))
			Expect(decoded["initial_prompt"]).To(ContainSubstring("Ada Lovelace"))
		})
	})

	Describe("formatTranscriptResult", func() {
		It("wraps the transcript in a JSON object", func() {
			res := formatTranscriptResult("some text")
			Expect(res).To(MatchJSON(`{"transcript": "some text"}`))
		})
	})
})
