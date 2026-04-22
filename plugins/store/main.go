//go:build wasip1
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(reqJSON), &probe); err != nil {
			return "", err
		}
		switch {
		case lo.HasKey(probe, "episode"):
			return handleRaw([]byte(reqJSON))
		case lo.HasKey(probe, "entry"):
			return handleStructured([]byte(reqJSON))
		default:
			return "", fmt.Errorf("request must contain 'episode' or 'entry'")
		}
	})
}

func handleRaw(in []byte) (string, error) {
	var req struct {
		OutputDir string `json:"output_dir"`
		Episode   struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Transcript string `json:"transcript"`
		} `json:"episode"`
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

func handleStructured(in []byte) (string, error) {
	var req struct {
		OutputDir string          `json:"output_dir"`
		Entry     json.RawMessage `json:"entry"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return "", err
	}
	var meta struct {
		EpisodeID string `json:"episode_id"`
	}
	_ = json.Unmarshal(req.Entry, &meta)

	path := fmt.Sprintf("%s/%s_entry.json", req.OutputDir, slug(meta.EpisodeID))
	pretty, _ := json.MarshalIndent(json.RawMessage(req.Entry), "", "  ")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, pretty, 0644); err != nil {
		return "", err
	}
	guest.LogMsg("stored entry: " + path)
	return fmt.Sprintf(`{"path":%q}`, path), nil
}

func main() {}
