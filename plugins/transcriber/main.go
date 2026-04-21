//go:build wasip1

// Plugin: transcriber
// Calls a configurable ASR HTTP endpoint via the host's generic http-post
// capability. Swap this plugin to change ASR engine (Whisper.cpp, Deepgram,
// AssemblyAI…) without touching the kernel.
package main

import (
	"encoding/json"

	"github.com/gavmor/wasm-microkernel/guest-bindings/plugin_world"
	host "github.com/gavmor/wasm-microkernel/guest-bindings/podpedia/kernel/host_capabilities"
	"github.com/samber/lo"
)

func main() {}

func init() { plugin_world.SetExportsPluginWorld(&TranscriberPlugin{}) }

type TranscriberPlugin struct{}

func (t *TranscriberPlugin) Execute(reqJSON string) (plugin_world.Result[string, string], error) {
	var req struct {
		AudioPath     string `json:"audio_path"`
		AudioURL      string `json:"audio_url"`
		TranscribeURL string `json:"transcribe_url"`
	}
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return plugin_world.Err[string, string]("bad request: " + err.Error()), nil
	}

	target, ok := lo.Coalesce(req.AudioPath, req.AudioURL)
	if !ok {
		return plugin_world.Err[string, string]("audio_path or audio_url required"), nil
	}

	if req.TranscribeURL == "" {
		host.LogMsg("no transcribe_url configured, skipping: " + target)
		return plugin_world.Ok[string, string](`{"transcript":""}`), nil
	}

	host.LogMsg("transcribing " + target)

	rawRes, err := host.HttpPost(req.TranscribeURL, buildTranscribeBody(target))
	if err != nil {
		return plugin_world.Err[string, string]("host http-post failed: " + err.Error()), nil
	}

	return plugin_world.Ok[string, string](formatTranscriptResult(parseASRResponse(rawRes))), nil
}
