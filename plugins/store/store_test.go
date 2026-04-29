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

		It("creates the correct entry file and meta sidecar", func() {
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

			metaPath := tmpDir + "/ep_123_meta.json"
			Expect(metaPath).To(BeAnExistingFile())
			metaBytes, err := os.ReadFile(metaPath)
			Expect(err).NotTo(HaveOccurred())
			var meta map[string]string
			Expect(json.Unmarshal(metaBytes, &meta)).To(Succeed())
			Expect(meta).To(HaveKey("audio_url"))
			Expect(meta).To(HaveKey("episode_id"))
			Expect(meta).To(HaveKey("title"))
			Expect(meta).To(HaveKey("pub_date"))
		})

		It("is idempotent — second call skips re-writing the entry", func() {
			reqJSON, _ := json.Marshal(req)

			_, err := HandleStructured(reqJSON)
			Expect(err).NotTo(HaveOccurred())

			entryPath := tmpDir + "/ep_123_my-scheme.json"
			beforeContent, err := os.ReadFile(entryPath)
			Expect(err).NotTo(HaveOccurred())

			// Modify the entry data — second call should NOT overwrite
			req["entry"] = map[string]any{"data": "different data"}
			reqJSON, _ = json.Marshal(req)
			_, err = HandleStructured(reqJSON)
			Expect(err).NotTo(HaveOccurred())

			afterContent, err := os.ReadFile(entryPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterContent).To(Equal(beforeContent))
		})

		It("backfills missing metadata even if the structured entry already exists (without overwriting)", func() {
			reqJSON, _ := json.Marshal(req)

			// 1. Initial write
			_, err := HandleStructured(reqJSON)
			Expect(err).NotTo(HaveOccurred())

			entryPath := tmpDir + "/ep_123_my-scheme.json"
			beforeContent, err := os.ReadFile(entryPath)
			Expect(err).NotTo(HaveOccurred())

			metaPath := tmpDir + "/ep_123_meta.json"
			Expect(metaPath).To(BeAnExistingFile())

			// 2. Delete only the metadata
			err = os.Remove(metaPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(metaPath).NotTo(BeAnExistingFile())

			// 3. Call again with DIFFERENT entry data - should NOT overwrite entry, but should restore meta
			req["entry"] = map[string]any{"data": "malicious/accidental overwrite"}
			reqJSON, _ = json.Marshal(req)
			_, err = HandleStructured(reqJSON)
			Expect(err).NotTo(HaveOccurred())

			// 4. Verify meta is back
			Expect(metaPath).To(BeAnExistingFile())

			// 5. Verify entry was NOT overwritten
			afterContent, err := os.ReadFile(entryPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterContent).To(Equal(beforeContent))
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
