package llm

import (
	"math/rand"
	"time"

	"github.com/gavmor/podpedia/internal/types"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) ExtractEntities(ep types.Episode) (types.EncyclopediaEntry, error) {
	// Simulate LLM inference time
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1500)+1000))

	return types.EncyclopediaEntry{
		EpisodeID: ep.ID,
		Guests: []types.Guest{
			{Name: "Jane Doe", Background: "AI Researcher", Ideology: "Open Source AI"},
		},
		Companies: []types.Company{
			{Name: "Acme Corp", BusinessModel: "B2B SaaS", Customers: "Enterprise Developers"},
		},
	}, nil
}
