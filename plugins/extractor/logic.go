package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gavmor/podpedia/internal/types"
)

// extractRequest is the wire contract with the host: it must stay
// structurally identical to the anonymous struct that
// internal/kernel.(*extractor).ExtractEntities marshals.
//
// types.Episode is resolved through the `replace github.com/gavmor/podpedia
// => ../..` in go.mod, so it is the same struct definition the host encodes
// with. Without that replace this module compiled against a published zip
// whose Episode carried no json tags, and the host's underscored keys
// (audio_url, pub_date) bound to nothing — Go's default field matching is
// case-insensitive but does not cross underscores.
type extractRequest struct {
	Episode types.Episode   `json:"episode"`
	Scheme  json.RawMessage `json:"scheme"`
}

// defaultScheme is the extraction schema this plugin uses when the host does
// not supply one of its own.
const defaultScheme = `{"guests":[{"name":"","background":"","ideology":""}],"companies":[{"name":"","business_model":"","customers":""}]}`

// resolveScheme picks the schema to put in front of the model and reports
// whether the host's scheme was rejected in favour of defaultScheme.
//
// A scheme that is absent, null, or not valid JSON is simply unset. A scheme
// that is a JSON *string* is a wire-format accident and is refused: the host
// once typed this field as []byte, which encoding/json renders as base64, so
// a scheme of `{}` arrived here as the literal `"e30="`. Prompting a model
// with that is strictly worse than prompting it with a schema this plugin
// knows is usable, and a host pinned to an older engine still sends it.
func resolveScheme(scheme json.RawMessage) (resolved string, rejected bool) {
	trimmed := strings.TrimSpace(string(scheme))
	switch {
	case trimmed == "" || trimmed == "null":
		return defaultScheme, false
	case strings.HasPrefix(trimmed, `"`), !json.Valid([]byte(trimmed)):
		return defaultScheme, true
	default:
		return trimmed, false
	}
}

func buildPrompt(title, content string, scheme json.RawMessage) string {
	schemaStr, _ := resolveScheme(scheme)

	return fmt.Sprintf("Extract data from this podcast episode.\n\nSchema:\n%s\n\nTitle: %s\nContent: %s",
		schemaStr, title, content)
}

func parseCompletion(raw string) (json.RawMessage, error) {
	raw = stripMarkdown(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON in completion")
	}

	js := raw[start : end+1]
	return json.RawMessage(js), nil
}

func stripMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) > 2 {
			return strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	return s
}
