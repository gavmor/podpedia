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
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return "", err
		}

		transcribeURL, _ := guest.Config("transcribe-url")
		if transcribeURL == "" {
			return "", fmt.Errorf("transcribe-url not configured")
		}

		target := req.AudioURL
		guest.LogMsg("transcribing: " + target)

		rawRes, err := guest.HttpPost(transcribeURL, buildTranscribeBody(target))
		if err != nil {
			return "", fmt.Errorf("host http-post failed: %v", err)
		}

		return formatTranscriptResult(parseASRResponse(rawRes)), nil
	})
}

func main() {}
