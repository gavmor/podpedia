//go:build wasip1

// Plugin: store
// Persists transcripts and encyclopedia entries via the host's file-write
// capability. Owns all path construction and serialization format decisions.
// Swap this plugin to change output format (markdown, SQLite, S3, etc.)
package main

import (
	"encoding/json"
	"fmt"

	"github.com/gavmor/wasm-microkernel/guest-bindings/plugin_world"
	host "github.com/gavmor/wasm-microkernel/guest-bindings/podpedia/kernel/host_capabilities"
	"github.com/samber/lo"
)

func main() {}

func init() { plugin_world.SetExportsPluginWorld(&StorePlugin{}) }

type StorePlugin struct{}

func (s *StorePlugin) Execute(reqJSON string) (plugin_world.Result[string, string], error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(reqJSON), &probe); err != nil {
		return plugin_world.Err[string, string]("bad request: " + err.Error()), nil
	}
	switch {
	case lo.HasKey(probe, "episode"):
		return handleRaw([]byte(reqJSON))
	case lo.HasKey(probe, "entry"):
		return handleStructured([]byte(reqJSON))
	default:
		return plugin_world.Err[string, string]("request must contain 'episode' or 'entry'"), nil
	}
}

func handleRaw(in []byte) (plugin_world.Result[string, string], error) {
	var req struct {
		OutputDir string `json:"output_dir"`
		Episode   struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Transcript string `json:"transcript"`
		} `json:"episode"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return plugin_world.Err[string, string]("bad raw request: " + err.Error()), nil
	}
	path := fmt.Sprintf("%s/%s_raw.txt", req.OutputDir, slug(req.Episode.ID))
	content, _ := lo.Coalesce(req.Episode.Transcript, fmt.Sprintf("# %s\n\n(no transcript)\n", req.Episode.Title))
	if err := host.FileWrite(path, content); err != nil {
		return plugin_world.Err[string, string]("host file-write failed: " + err.Error()), nil
	}
	host.LogMsg("stored raw: " + path)
	return plugin_world.Ok[string, string](fmt.Sprintf(`{"path":%q}`, path)), nil
}

func handleStructured(in []byte) (plugin_world.Result[string, string], error) {
	var req struct {
		OutputDir string          `json:"output_dir"`
		Entry     json.RawMessage `json:"entry"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return plugin_world.Err[string, string]("bad structured request: " + err.Error()), nil
	}
	var meta struct {
		EpisodeID string `json:"episode_id"`
	}
	_ = json.Unmarshal(req.Entry, &meta)

	path := fmt.Sprintf("%s/%s_entry.json", req.OutputDir, slug(meta.EpisodeID))
	pretty, _ := json.MarshalIndent(json.RawMessage(req.Entry), "", "  ")
	if err := host.FileWrite(path, string(pretty)); err != nil {
		return plugin_world.Err[string, string]("host file-write failed: " + err.Error()), nil
	}
	host.LogMsg("stored entry: " + path)
	return plugin_world.Ok[string, string](fmt.Sprintf(`{"path":%q}`, path)), nil
}
