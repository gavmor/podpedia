package main

import (
	"encoding/json"
	"fmt"

	"github.com/gavmor/wasm-microkernel/guest"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var req struct {
			AudioURL string `json:"audio_url"`
			Prompt   string `json:"prompt"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return "", err
		}

		transcribeURL, _ := guest.Config("transcribe-url")
		if transcribeURL == "" {
			return "", fmt.Errorf("transcribe-url not configured")
		}

		guest.LogMsg("transcribing: " + req.AudioURL)

		rawRes, err := guest.HttpPost(transcribeURL, buildTranscribeBody(req.AudioURL, req.Prompt))
		if err != nil {
			return "", fmt.Errorf("host http-post failed: %v", err)
		}

		return formatTranscriptResult(parseASRResponse(rawRes)), nil
	})
}

func main() {}
