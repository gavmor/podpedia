//go:build wasip1

// Plugin: transcriber
// Delegates transcription to the host's `transcribe` capability.
// Swap this plugin to change ASR engine (Whisper, Deepgram, AssemblyAI…)
// without touching the kernel.
package main

import (
	"encoding/json"
	"unsafe"

	"github.com/gavmor/wasm-microkernel/abi"
)

func main() {}

//go:wasmimport podpedia_host transcribe
func hostTranscribe(fatPtr uint64) uint64

//go:wasmimport podpedia_host log
func hostLog(fatPtr uint64)

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(input []byte) []byte {
		var req struct {
			AudioPath string `json:"audio_path"`
			AudioURL  string `json:"audio_url"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}
		target := req.AudioPath
		if target == "" {
			target = req.AudioURL
		}
		if target == "" {
			return errBytes("audio_path or audio_url required")
		}

		logMsg("transcribing " + target)
		pathJSON, _ := json.Marshal(target)
		result := hostTranscribe(abi.ReturnBytes(pathJSON))

		off, ln := abi.DecodeFatPointer(result)
		raw := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(off))), ln)
		var transcript string
		if err := json.Unmarshal(raw, &transcript); err != nil {
			transcript = string(raw)
		}

		out, _ := json.Marshal(map[string]string{"transcript": transcript})
		return out
	})
}

func logMsg(s string) { hostLog(abi.ReturnBytes([]byte(s))) }
func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
