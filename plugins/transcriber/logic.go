package main

import (
	"encoding/json"
	"fmt"
)

func parseASRResponse(rawRes string) string {
	var resp struct {
		Transcript string `json:"transcript"`
	}
	if err := json.Unmarshal([]byte(rawRes), &resp); err != nil {
		return rawRes
	}
	return resp.Transcript
}

func buildTranscribeBody(audioTarget string) string {
	body, _ := json.Marshal(map[string]string{"audio_url": audioTarget})
	return string(body)
}

func formatTranscriptResult(transcript string) string {
	return fmt.Sprintf(`{"transcript":%q}`, transcript)
}
