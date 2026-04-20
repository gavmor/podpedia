//go:build wasip1

// Plugin: transcriber
// Delegates transcription to an ASR HTTP endpoint via the host's generic
// http_post capability. Swap this plugin to change ASR engine (Whisper.cpp
// HTTP server, Deepgram, AssemblyAI…) without touching the kernel.
package main

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/samber/lo"
)

func main() {}

//go:wasmimport podpedia_host http_post
func hostHTTPPost(fatPtr uint64) uint64

//go:wasmimport podpedia_host log
func hostLog(fatPtr uint64)

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
			logMsg("no transcribe_url configured, skipping: " + target)
			return []byte(`{"transcript":""}`)
		}

		logMsg("transcribing " + target)
		transcript := callTranscribeAPI(req.TranscribeURL, target)
		return []byte(fmt.Sprintf(`{"transcript":%q}`, transcript))
	})
}

// callTranscribeAPI POSTs {audio_url} to the configured transcription endpoint
// and returns the transcript string. The plugin owns the request format — swap
// this function to speak Deepgram, AssemblyAI, or a local Whisper HTTP server.
func callTranscribeAPI(transcribeURL, audioTarget string) string {
	reqBody, _ := json.Marshal(map[string]string{"audio_url": audioTarget})
	postReq, _ := json.Marshal(map[string]string{
		"url":  transcribeURL,
		"body": string(reqBody),
	})
	result := hostHTTPPost(abi.ReturnBytes(postReq))
	off, ln := abi.DecodeFatPointer(result)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(off))), ln)
	var resp struct {
		Transcript string `json:"transcript"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return string(raw)
	}
	return resp.Transcript
}

func logMsg(s string) { hostLog(abi.ReturnBytes([]byte(s))) }
func errBytes(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error":%q}`, msg))
}
