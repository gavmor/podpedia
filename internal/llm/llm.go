package llm

import "fmt"

func ExtractEntities(transcript string) (string, error) {
	fmt.Printf("[LLM] Extracting entities from transcript...\n")
	return fmt.Sprintf("Fake structured entities from: %s", transcript), nil
}
