package main

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extractor Plugin Logic", func() {
	Describe("buildPrompt", func() {
		It("contains the episode title and transcript content", func() {
			prompt := buildPrompt("My Episode", "Some transcript content", nil)
			Expect(prompt).To(ContainSubstring("My Episode"))
			Expect(prompt).To(ContainSubstring("Some transcript content"))
		})

		Context("with a custom scheme", func() {
			It("includes the custom scheme in the prompt", func() {
				scheme := json.RawMessage(`{"ideology":""}`)
				prompt := buildPrompt("Title", "Content", scheme)
				Expect(prompt).To(ContainSubstring(`"ideology"`))
			})
		})
	})

	Describe("parseCompletion", func() {
		Context("with valid JSON completion", func() {
			It("correctly parses the JSON", func() {
				raw := `{"guests":[{"name":"Alice","background":"Engineer","ideology":"pragmatic"}],"companies":[{"name":"Acme","business_model":"SaaS","customers":"SMBs"}]}`
				result, err := parseCompletion(raw)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(MatchJSON(raw))
			})
		})

		Context("with JSON embedded in a preamble", func() {
			It("extracts and parses the JSON", func() {
				rawJSON := `{"guests":[],"companies":[{"name":"Beta","business_model":"B2B","customers":"enterprises"}]}`
				raw := "Here is the extracted data: " + rawJSON
				result, err := parseCompletion(raw)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(MatchJSON(rawJSON))
			})
		})

		Context("when the completion is invalid", func() {
			It("returns an error if no JSON is found", func() {
				_, err := parseCompletion("no json here at all")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no JSON in completion"))
			})

			It("returns an error for an empty string", func() {
				_, err := parseCompletion("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no JSON in completion"))
			})

			It("returns an error for malformed JSON", func() {
				_, err := parseCompletion(`{"guests": [broken`)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
