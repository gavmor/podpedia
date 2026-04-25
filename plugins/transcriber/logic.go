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
// title and notes are forwarded as initial_prompt so the model biases
// towards correct spelling of proper nouns mentioned in the episode metadata.
func buildTranscribeBody(audioTarget, title, notes string) string {
	prompt := buildPrompt(title, notes)
	fields := map[string]string{"audio_url": audioTarget}
	if prompt != "" {
		fields["initial_prompt"] = prompt
	}
	body, _ := json.Marshal(fields)
	return string(body)
}

func buildPrompt(title, notes string) string {
	switch {
	case title != "" && notes != "":
		return "Episode: " + title + "\n\n" + notes
	case title != "":
		return "Episode: " + title
	default:
		return notes
	}
}

func formatTranscriptResult(transcript string) string {
	return fmt.Sprintf(`{"transcript":%q}`, transcript)
}
