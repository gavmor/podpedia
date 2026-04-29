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

// HandleRaw stores the transcription text for an episode.
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

	if err := atomicWrite(path, []byte(content)); err != nil {
		return "", fmt.Errorf("failed to write raw data: %w", err)
	}

	guest.LogMsg("stored raw: " + path)
	return fmt.Sprintf(`{"path":%q}`, path), nil
}

// HandleStructured stores the extracted structured data for an episode,
// along with a metadata sidecar file for downstream indexing.
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

	id := req.Episode.ID
	if id == "" {
		id = req.Episode.AudioURL
	}
	entryID := slug(id)

	// Idempotency: skip only if both the structured entry and meta sidecar exist
	structuredPath := fmt.Sprintf("%s/%s_%s.json", req.OutputDir, entryID, req.SchemeID)
	metaPath := fmt.Sprintf("%s/%s_meta.json", req.OutputDir, entryID)

	structuredExists := false
	if _, err := os.Stat(structuredPath); err == nil {
		structuredExists = true
	}
	metaExists := false
	if _, err := os.Stat(metaPath); err == nil {
		metaExists = true
	}

	if structuredExists && metaExists {
		guest.LogMsg("skipping store: " + structuredPath)
		return fmt.Sprintf(`{"path":%q}`, structuredPath), nil
	}

	// Write structured data if it doesn't exist
	if !structuredExists {
		pretty, _ := json.MarshalIndent(req.Entry, "", "  ")
		if err := atomicWrite(structuredPath, pretty); err != nil {
			return "", fmt.Errorf("failed to write structured data: %w", err)
		}
	}

	// Write metadata sidecar if it doesn't exist
	if !metaExists {
		meta := struct {
			EpisodeID string `json:"episode_id"`
			AudioURL  string `json:"audio_url"`
			Title     string `json:"title"`
			PubDate   string `json:"pub_date"`
		}{
			EpisodeID: req.Episode.ID,
			AudioURL:  req.Episode.AudioURL,
			Title:     req.Episode.Title,
			PubDate:   req.Episode.PubDate,
		}
		metaJSON, _ := json.MarshalIndent(meta, "", "  ")
		if err := atomicWrite(metaPath, metaJSON); err != nil {
			return "", fmt.Errorf("failed to write metadata sidecar: %w", err)
		}
	}

	guest.LogMsg("stored structured: " + structuredPath)
	return fmt.Sprintf(`{"path":%q}`, structuredPath), nil
}

// atomicWrite creates a temporary file, writes content to it, and then
// atomically renames it to the final destination. This prevents data
// corruption from partial writes.
func atomicWrite(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Use CreateTemp to avoid race conditions with concurrent writers
	tmpFile, err := os.CreateTemp(dir, "atomic-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions to 0644 (world-readable) before rename
	if err := os.Chmod(tmpFile.Name(), 0644); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Sync parent directory to ensure the rename is durable
	df, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("failed to open directory for sync: %w", err)
	}
	defer df.Close()
	if err := df.Sync(); err != nil {
		return fmt.Errorf("failed to sync directory: %w", err)
	}

	return nil
}

// The slug function is now in logic.go.

func main() {}
