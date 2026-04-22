package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(reqJSON), &probe); err != nil {
			return "", err
		}
		if _, ok := probe["entry"]; ok {
			return HandleStructured([]byte(reqJSON))
		}
		if _, ok := probe["episode"]; ok {
			return HandleRaw([]byte(reqJSON))
		}
		return "", fmt.Errorf("request must contain 'episode' and optionally 'entry'")
	})
}

func HandleRaw(in []byte) (string, error) {
	var req struct {
		OutputDir string        `json:"output_dir"`
		Episode   types.Episode `json:"episode"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/%s_raw.txt", req.OutputDir, slug(req.Episode.ID))
	content, _ := lo.Coalesce(req.Episode.Transcript, fmt.Sprintf("# %s\n\n(no transcript)\n", req.Episode.Title))

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	guest.LogMsg("stored raw: " + path)
	return fmt.Sprintf(`{"path":%q}`, path), nil
}

func HandleStructured(in []byte) (string, error) {
	var req struct {
		OutputDir string          `json:"output_dir"`
		Episode   types.Episode   `json:"episode"`
		Entry     json.RawMessage `json:"entry"`
		SchemeID  string          `json:"scheme_id"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return "", err
	}

	suffix := "entry"
	if req.SchemeID != "" {
		suffix = req.SchemeID
	}

	id := req.Episode.ID
	if id == "" {
		id = "unknown"
	}

	path := fmt.Sprintf("%s/%s_%s.json", req.OutputDir, slug(id), suffix)
	pretty, _ := json.MarshalIndent(req.Entry, "", "  ")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, pretty, 0644); err != nil {
		return "", err
	}
	guest.LogMsg("stored structured: " + path)
	return fmt.Sprintf(`{"path":%q}`, path), nil
}

func main() {}
