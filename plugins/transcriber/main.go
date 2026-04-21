//go:build wasip1

// Plugin: transcriber
// Calls a configurable ASR HTTP endpoint via the host's generic http-post
// capability. Swap this plugin to change ASR engine (Whisper.cpp, Deepgram,
// AssemblyAI…) without touching the kernel.
package main

import (
	"encoding/json"

	hostcapabilities "github.com/gavmor/podpedia/gen/podpedia/kernel/host-capabilities"
	pluginworld "github.com/gavmor/podpedia/gen/podpedia/kernel/plugin-world"
	"github.com/samber/lo"
	"go.bytecodealliance.org/cm"
)

func main() {}

// Result is a convenience alias used throughout this file.
type Result = cm.Result[string, string, string]

func ok(s string) Result  { return cm.OK[Result](s) }
func fail(s string) Result { return cm.Err[Result](s) }

func init() {
	pluginworld.Exports.Execute = func(reqJSON string) Result {
		var req struct {
			AudioPath     string `json:"audio_path"`
			AudioURL      string `json:"audio_url"`
			TranscribeURL string `json:"transcribe_url"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return fail("bad request: " + err.Error())
		}

		target, hasTarget := lo.Coalesce(req.AudioPath, req.AudioURL)
		if !hasTarget {
			return fail("audio_path or audio_url required")
		}

		if req.TranscribeURL == "" {
			hostcapabilities.LogMsg("no transcribe_url configured, skipping: " + target)
			return ok(`{"transcript":""}`)
		}

		hostcapabilities.LogMsg("transcribing " + target)

		r := hostcapabilities.HTTPPost(req.TranscribeURL, buildTranscribeBody(target))
		if r.IsErr() {
			return fail("host http-post failed: " + *r.Err())
		}

		return ok(formatTranscriptResult(parseASRResponse(*r.OK())))
	}
}
