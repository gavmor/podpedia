//go:build wasip1

// Plugin: store
// Persists transcripts and encyclopedia entries via the host's file_write
// capability. Owns all path construction and serialization format decisions.
// Swap this plugin to change output format (markdown, SQLite, S3, etc.)
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
)

func main() {}

//go:wasmimport podpedia_host file_write
func hostFileWrite(fatPtr uint64) uint32

//go:wasmimport podpedia_host log
func hostLog(fatPtr uint64)

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(input []byte) []byte {
		// Dispatch on which fields are present.
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(input, &probe); err != nil {
			return errBytes("bad request: " + err.Error())
		}
		if _, ok := probe["episode"]; ok {
			return handleRaw(input)
		}
		if _, ok := probe["entry"]; ok {
			return handleStructured(input)
		}
		return errBytes("request must contain 'episode' or 'entry'")
	})
}

func handleRaw(input []byte) []byte {
	var req struct {
		OutputDir string `json:"output_dir"`
		Episode   struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Transcript string `json:"transcript"`
		} `json:"episode"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return errBytes("bad raw request: " + err.Error())
	}
	path := fmt.Sprintf("%s/%s_raw.txt", req.OutputDir, slug(req.Episode.ID))
	content := req.Episode.Transcript
	if content == "" {
		content = "# " + req.Episode.Title + "\n\n(no transcript)\n"
	}
	if !writeFile(path, content) {
		return errBytes("file_write failed: " + path)
	}
	logMsg("stored raw: " + path)
	out, _ := json.Marshal(map[string]string{"path": path})
	return out
}

func handleStructured(input []byte) []byte {
	var req struct {
		OutputDir string          `json:"output_dir"`
		Entry     json.RawMessage `json:"entry"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return errBytes("bad structured request: " + err.Error())
	}

	// Extract episode_id from the entry for the filename.
	var meta struct {
		EpisodeID string `json:"episode_id"`
	}
	json.Unmarshal(req.Entry, &meta)

	path := fmt.Sprintf("%s/%s_entry.json", req.OutputDir, slug(meta.EpisodeID))
	pretty, _ := json.MarshalIndent(json.RawMessage(req.Entry), "", "  ")
	if !writeFile(path, string(pretty)) {
		return errBytes("file_write failed: " + path)
	}
	logMsg("stored entry: " + path)
	out, _ := json.Marshal(map[string]string{"path": path})
	return out
}

func writeFile(path, data string) bool {
	payload, _ := json.Marshal(map[string]string{"path": path, "data": data})
	return hostFileWrite(abi.ReturnBytes(payload)) == 1
}

func slug(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-':
			b.WriteRune(c)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func logMsg(s string) { hostLog(abi.ReturnBytes([]byte(s))) }
func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
