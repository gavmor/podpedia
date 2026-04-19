package pipeline

import (
	"fmt"
	"github.com/gavmor/podpedia/internal/llm"
	"github.com/gavmor/podpedia/internal/storage"
	"github.com/gavmor/podpedia/internal/transcription"
)

func Run(rssURL string) error {
	fmt.Printf("[Pipeline] Starting for feed: %s\n", rssURL)

	// Mocking a list of episodes found in RSS
	episodes := []string{"ep1", "ep2", "ep3"}

	for _, ep := range episodes {
		fmt.Printf("[Pipeline] Processing %s\n", ep)

		transcript, err := transcription.Transcribe(ep)
		if err != nil {
			return err
		}

		entities, err := llm.ExtractEntities(transcript)
		if err != nil {
			return err
		}

		err = storage.Save(ep, entities)
		if err != nil {
			return err
		}
	}

	fmt.Println("[Pipeline] Pipeline completed successfully.")
	return nil
}
