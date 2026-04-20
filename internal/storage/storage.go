package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gavmor/podpedia/internal/types"
)

func SaveRawData(outputDir string, ep types.Episode) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	fmt.Printf("[Storage] Saving raw data for %s\n", ep.ID)
	return os.WriteFile(fmt.Sprintf("%s/%s_raw.txt", outputDir, ep.ID), []byte(ep.Transcript), 0644)
}

func SaveStructuredData(outputDir string, entry types.EncyclopediaEntry) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	fmt.Printf("[Storage] Saving structured data for %s\n", entry.EpisodeID)
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(fmt.Sprintf("%s/%s_entry.json", outputDir, entry.EpisodeID), jsonData, 0644)
}
