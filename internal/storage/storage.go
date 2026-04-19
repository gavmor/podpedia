package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gavmor/podpedia/internal/types"
)

func SaveRawData(ep types.Episode) error {
	fmt.Printf("[Storage] Saving raw data for %s\n", ep.ID)
	return os.WriteFile(fmt.Sprintf("%s_raw.txt", ep.ID), []byte(ep.Transcript), 0644)
}

func SaveStructuredData(entry types.EncyclopediaEntry) error {
	fmt.Printf("[Storage] Saving structured data for %s\n", entry.EpisodeID)
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(fmt.Sprintf("%s_entry.json", entry.EpisodeID), jsonData, 0644)
}
