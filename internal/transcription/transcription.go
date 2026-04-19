package transcription

import "fmt"

func Transcribe(episodeID string) (string, error) {
	fmt.Printf("[Transcription] Transcribing episode: %s\n", episodeID)
	return fmt.Sprintf("Fake transcript for %s", episodeID), nil
}
