package transcription

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Transcribe(audioURL string) (string, error) {
	// Simulate processing time
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)+500))
	return fmt.Sprintf("Raw transcript for %s: This is a podcast episode discussing AI and SaaS models...", audioURL), nil
}
