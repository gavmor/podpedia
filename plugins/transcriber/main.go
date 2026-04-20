//go:build wasip1

// Plugin: transcriber
// Calls a configurable ASR HTTP endpoint via the host's generic http-post
// capability. Swap this plugin to change ASR engine (Whisper.cpp, Deepgram,
// AssemblyAI…) without touching the kernel.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func main() {}

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(input []byte) []byte {
		var req struct {
			AudioPath     string `json:"audio_path"`
			AudioURL      string `json:"audio_url"`
			TranscribeURL string `json:"transcribe_url"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}

		target, ok := lo.Coalesce(req.AudioPath, req.AudioURL)
		if !ok {
			return errBytes("audio_path or audio_url required")
		}

		if req.TranscribeURL == "" {
			guest.Log("no transcribe_url configured, skipping: " + target)
			return []byte(`{"transcript":""}`)
		}

		guest.Log("transcribing " + target)

		reqBody, _ := json.Marshal(map[string]string{"audio_url": target})
		raw := guest.HTTPPost(req.TranscribeURL, string(reqBody))

		var resp struct {
			Transcript string `json:"transcript"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return []byte(fmt.Sprintf(`{"transcript":%q}`, string(raw)))
		}
		return []byte(fmt.Sprintf(`{"transcript":%q}`, resp.Transcript))
	})
}

func errBytes(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error":%q}`, msg))
}
