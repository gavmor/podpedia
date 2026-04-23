package main

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store Plugin Logic", func() {
	Describe("slug", func() {
		DescribeTable("slugification cases",
			func(input, expected string) {
				Expect(slug(input)).To(Equal(expected))
			},
			Entry("alphanumeric pass-through", "hello", "hello"),
			Entry("mixed case and numbers", "Hello123", "Hello123"),
			Entry("hyphen to underscore", "abc-def", "abc_def"),
			Entry("all caps", "ABC", "ABC"),
			Entry("non-alphanumeric to underscore", "http://id/123", "http___id_123"),
		)
	})

	Describe("HandleStructured", func() {
		var (
			tmpDir  string
			episode map[string]any
			entry   map[string]any
			req     map[string]any
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "store-test")
			Expect(err).NotTo(HaveOccurred())

			episode = map[string]any{
				"id": "ep-123",
			}
			entry = map[string]any{
				"data": "some data",
			}
			req = map[string]any{
				"output_dir": tmpDir,
				"episode":    episode,
				"entry":      entry,
				"scheme_id":  "my-scheme",
			}
		})

		AfterEach(func() {
			_ = os.RemoveAll(tmpDir)
		})

		It("creates the correct filename", func() {
			reqJSON, _ := json.Marshal(req)

			resJSON, err := HandleStructured(reqJSON)
			Expect(err).NotTo(HaveOccurred())

			var res struct {
				Path string `json:"path"`
			}
			err = json.Unmarshal([]byte(resJSON), &res)
			Expect(err).NotTo(HaveOccurred())

			Expect(res.Path).To(HaveSuffix("ep_123_my-scheme.json"))
			Expect(res.Path).To(BeAnExistingFile())
		})

		Context("with non-alphanumeric characters in the ID", func() {
			BeforeEach(func() {
				episode["id"] = "http://id/123"
				req["scheme_id"] = "test"
			})

			It("slugifies the ID in the filename", func() {
				reqJSON, _ := json.Marshal(req)

				resJSON, err := HandleStructured(reqJSON)
				Expect(err).NotTo(HaveOccurred())

				var res struct {
					Path string `json:"path"`
				}
				_ = json.Unmarshal([]byte(resJSON), &res)

				Expect(res.Path).To(ContainSubstring("http___id_123_test.json"))
			})
		})
	})
})
