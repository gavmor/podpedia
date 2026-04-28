package podpedia_test

import (
	"testing"

	"github.com/gavmor/podpedia/pkg/podpedia"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

func TestPodpedia(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Podpedia Package Suite")
}

var _ = Describe("App", func() {
	It("can be initialized with functional options", func() {
		fs := afero.NewMemMapFs()
		app, err := podpedia.NewApp(
			podpedia.WithWorkspace(fs),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(app).NotTo(BeNil())
	})
})
