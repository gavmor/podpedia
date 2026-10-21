package kernel

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gavmor/podpedia/internal/types"
)

// goldenExtractRequest is the exact JSON the host hands to plugins/extractor
// for the episode and scheme below.
//
// It is duplicated, byte for byte, as hostRequestWithSchemeJSON in
// plugins/extractor/request_test.go. The plugin is a separate Go module, so
// there is no shared type the two sides can both assert against; a golden
// literal on each side is what turns a wire-format drift into two failing
// tests instead of a silent runtime divergence.
const goldenExtractRequest = `{"episode":{` +
	`"id":"fx-ep-02",` +
	`"title":"Fixture Episode Two",` +
	`"audio_url":"https://example.test/fx-ep-02.mp3",` +
	`"pub_date":"Mon, 20 Oct 2026 09:00:00 +0000",` +
	`"description":"A short description.",` +
	`"author":"Fixture Author",` +
	`"duration":"00:12:34",` +
	`"explicit":true,` +
	`"categories":["Technology"],` +
	`"transcript":"Welcome to Fixture Cast."` +
	`},"scheme":{"guests":[{"name":""}]}}`

func goldenEpisode() types.Episode {
	return types.Episode{
		ID:          "fx-ep-02",
		Title:       "Fixture Episode Two",
		AudioURL:    "https://example.test/fx-ep-02.mp3",
		PubDate:     "Mon, 20 Oct 2026 09:00:00 +0000",
		Description: "A short description.",
		Author:      "Fixture Author",
		Duration:    "00:12:34",
		Explicit:    true,
		Categories:  []string{"Technology"},
		Transcript:  "Welcome to Fixture Cast.",
	}
}

// TestExtractRequestCarriesTheSchemeAsJSON is the regression pin for the
// defect that made this file necessary: the request struct typed Scheme as
// []byte, and encoding/json renders a []byte as a base64 *string*. A scheme
// of `{}` therefore arrived at the plugin as the seven-byte literal `"e30="`,
// which the plugin pasted into the LLM prompt verbatim.
func TestExtractRequestCarriesTheSchemeAsJSON(t *testing.T) {
	ep := goldenEpisode()
	wire, err := json.Marshal(newExtractRequest(&ep, json.RawMessage(`{"guests":[{"name":""}]}`)))
	if err != nil {
		t.Fatalf("marshal extract request: %v", err)
	}

	if got := string(wire); got != goldenExtractRequest {
		t.Errorf("extract request wire format changed.\n got: %s\nwant: %s", got, goldenExtractRequest)
	}

	// The failure mode stated as an invariant, so the reason this test exists
	// survives a future rewrite of the golden literal above.
	var envelope struct {
		Scheme json.RawMessage `json:"scheme"`
	}
	if err := json.Unmarshal(wire, &envelope); err != nil {
		t.Fatalf("unmarshal extract request: %v", err)
	}
	if strings.HasPrefix(string(envelope.Scheme), `"`) {
		t.Errorf("scheme reached the wire as a JSON string (%s), not as JSON. "+
			"A []byte-typed field base64-encodes; the plugin uses the quoted text as its schema.",
			envelope.Scheme)
	}
}

// TestExtractRequestOmitsAnAbsentSchemeAsNull pins the other half of the
// plugin's contract: plugins/extractor falls back to its own schema when the
// host sends none, and it recognises that case by the literal `null`.
func TestExtractRequestOmitsAnAbsentSchemeAsNull(t *testing.T) {
	ep := goldenEpisode()
	wire, err := json.Marshal(newExtractRequest(&ep, nil))
	if err != nil {
		t.Fatalf("marshal extract request: %v", err)
	}

	if !strings.HasSuffix(string(wire), `,"scheme":null}`) {
		t.Errorf("an absent scheme must reach the plugin as null, got: %s", wire)
	}
}
