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

// buildTranscribeBody constructs the JSON POST body for the ASR endpoint.
// prompt is forwarded as initial_prompt so the model biases towards correct
// spelling of proper nouns. The caller is responsible for constructing the
// prompt from whatever domain context it has.
func buildTranscribeBody(audioTarget, prompt string) string {
	fields := map[string]string{"audio_url": audioTarget}
	if prompt != "" {
		fields["initial_prompt"] = prompt
	}
	body, _ := json.Marshal(fields)
	return string(body)
}

func formatTranscriptResult(transcript string) string {
	return fmt.Sprintf(`{"transcript":%q}`, transcript)
}
