//go:build wasip1

// Plugin: store
// Persists transcripts and encyclopedia entries via the host's file-write
// capability. Owns all path construction and serialization format decisions.
// Swap this plugin to change output format (markdown, SQLite, S3, etc.)
package main

import (
	"encoding/json"
	"fmt"

	hostcapabilities "github.com/gavmor/podpedia/gen/podpedia/kernel/host-capabilities"
	pluginworld "github.com/gavmor/podpedia/gen/podpedia/kernel/plugin-world"
	"github.com/samber/lo"
	"go.bytecodealliance.org/cm"
)

func main() {}

// Result is a convenience alias used throughout this file.
type Result = cm.Result[string, string, string]

func ok(s string) Result  { return cm.OK[Result](s) }
func fail(s string) Result { return cm.Err[Result](s) }

func init() {
	pluginworld.Exports.Execute = func(reqJSON string) Result {
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(reqJSON), &probe); err != nil {
			return fail("bad request: " + err.Error())
		}
		switch {
		case lo.HasKey(probe, "episode"):
			return handleRaw([]byte(reqJSON))
		case lo.HasKey(probe, "entry"):
			return handleStructured([]byte(reqJSON))
		default:
			return fail("request must contain 'episode' or 'entry'")
		}
	}
}

func handleRaw(in []byte) Result {
	var req struct {
		OutputDir string `json:"output_dir"`
		Episode   struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Transcript string `json:"transcript"`
		} `json:"episode"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return fail("bad raw request: " + err.Error())
	}
	path := fmt.Sprintf("%s/%s_raw.txt", req.OutputDir, slug(req.Episode.ID))
	content, _ := lo.Coalesce(req.Episode.Transcript, fmt.Sprintf("# %s\n\n(no transcript)\n", req.Episode.Title))
	r := hostcapabilities.FileWrite(path, content)
	if r.IsErr() {
		return fail("host file-write failed: " + *r.Err())
	}
	hostcapabilities.LogMsg("stored raw: " + path)
	return ok(fmt.Sprintf(`{"path":%q}`, path))
}

func handleStructured(in []byte) Result {
	var req struct {
		OutputDir string          `json:"output_dir"`
		Entry     json.RawMessage `json:"entry"`
	}
	if err := json.Unmarshal(in, &req); err != nil {
		return fail("bad structured request: " + err.Error())
	}
	var meta struct {
		EpisodeID string `json:"episode_id"`
	}
	_ = json.Unmarshal(req.Entry, &meta)

	path := fmt.Sprintf("%s/%s_entry.json", req.OutputDir, slug(meta.EpisodeID))
	pretty, _ := json.MarshalIndent(json.RawMessage(req.Entry), "", "  ")
	r := hostcapabilities.FileWrite(path, string(pretty))
	if r.IsErr() {
		return fail("host file-write failed: " + *r.Err())
	}
	hostcapabilities.LogMsg("stored entry: " + path)
	return ok(fmt.Sprintf(`{"path":%q}`, path))
}
