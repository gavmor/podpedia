//go:build wasip1

// Plugin: store
// Persists transcripts and encyclopedia entries via the host's file-write
// capability. Owns all path construction and serialization format decisions.
// Swap this plugin to change output format (markdown, SQLite, S3, etc.)
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func main() {}

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(in []byte) []byte {
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(in, &probe); err != nil {
			return errBytes("bad request: " + err.Error())
		}
		switch {
		case lo.HasKey(probe, "episode"):
			return handleRaw(in)
		case lo.HasKey(probe, "entry"):
			return handleStructured(in)
		default:
			return errBytes("request must contain 'episode' or 'entry'")
		}
	})
}

func handleRaw(in []byte) []byte {
	var req struct {
		OutputDir string `json:"output_dir"`
		Episode   struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Transcript string `json:"transcript"`
		} `json:"episode"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return errBytes("bad raw request: " + err.Error())
	}
	path := fmt.Sprintf("%s/%s_raw.txt", req.OutputDir, slug(req.Episode.ID))
	content, _ := lo.Coalesce(req.Episode.Transcript, fmt.Sprintf("# %s\n\n(no transcript)\n", req.Episode.Title))
	if !guest.FileWrite(path, content) {
		return errBytes("file_write failed: " + path)
	}
	guest.Log("stored raw: " + path)
	return []byte(fmt.Sprintf(`{"path":%q}`, path))
}

func handleStructured(in []byte) []byte {
	var req struct {
		OutputDir string          `json:"output_dir"`
		Entry     json.RawMessage `json:"entry"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return errBytes("bad structured request: " + err.Error())
	}
	var meta struct {
		EpisodeID string `json:"episode_id"`
	}
	_ = json.Unmarshal(req.Entry, &meta)

	path := fmt.Sprintf("%s/%s_entry.json", req.OutputDir, slug(meta.EpisodeID))
	pretty, _ := json.MarshalIndent(json.RawMessage(req.Entry), "", "  ")
	if !guest.FileWrite(path, string(pretty)) {
		return errBytes("file_write failed: " + path)
	}
	guest.Log("stored entry: " + path)
	return []byte(fmt.Sprintf(`{"path":%q}`, path))
}

func slug(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			return r
		default:
			return '_'
		}
	}, s)
}

func errBytes(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error":%q}`, msg))
}
