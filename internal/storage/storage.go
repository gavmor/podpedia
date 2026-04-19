package storage

import "fmt"

func Save(episodeID string, data string) error {
	fmt.Printf("[Storage] Saving data for episode %s: %s\n", episodeID, data)
	return nil
}
