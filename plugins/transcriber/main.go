package main

import (
	"encoding/json"
	"fmt"

	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var req struct {
			AudioPath     string `json:"audio_path"`
			AudioURL      string `json:"audio_url"`
			TranscribeURL string `json:"transcribe_url"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return "", err
		}

		target, hasTarget := lo.Coalesce(req.AudioPath, req.AudioURL)
		if !hasTarget {
			return "", fmt.Errorf("audio_path or audio_url required")
		}

		if req.TranscribeURL == "" {
			guest.LogMsg("no transcribe_url configured, skipping: " + target)
			return `{"transcript":""}`, nil
		}

		guest.LogMsg("transcribing " + target)

		rawRes, err := guest.HttpPost(req.TranscribeURL, buildTranscribeBody(target))
		if err != nil {
			return "", fmt.Errorf("host http-post failed: %v", err)
		}

		return formatTranscriptResult(parseASRResponse(rawRes)), nil
	})
}

func main() {}
