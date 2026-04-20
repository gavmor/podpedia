//go:build wasip1

// Plugin: rss
// Receives raw RSS XML from the host (which fetched the feed URL) and returns
// a structured list of podcast episodes as JSON. No host imports needed —
// this plugin is pure computation.
package main

import (
	"encoding/json"
	"encoding/xml"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
)

func main() {}

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(input []byte) []byte {
		var req struct {
			XML string `json:"xml"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}
		p, eps, err := parseRSS(req.XML)
		if err != nil {
			return errBytes("rss parse: " + err.Error())
		}
		out, _ := json.Marshal(map[string]any{"podcast": p, "episodes": eps})
		return out
	})
}

// ── RSS structs ───────────────────────────────────────────────────────────────

type xmlDoc struct {
	Channel xmlChannel `xml:"channel"`
}
type xmlChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []xmlItem `xml:"item"`
}
type xmlItem struct {
	GUID        string       `xml:"guid"`
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	PubDate     string       `xml:"pubDate"`
	Enclosure   xmlEnclosure `xml:"enclosure"`
	Duration    string       `xml:"duration"`
}
type xmlEnclosure struct {
	URL string `xml:"url,attr"`
}

type outPodcast struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
type outEpisode struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PubDate     string `json:"pub_date"`
	AudioURL    string `json:"audio_url"`
	Duration    string `json:"duration"`
}

func parseRSS(raw string) (outPodcast, []outEpisode, error) {
	var doc xmlDoc
	if err := xml.NewDecoder(strings.NewReader(raw)).Decode(&doc); err != nil {
		return outPodcast{}, nil, err
	}
	ch := doc.Channel
	p := outPodcast{Title: ch.Title, Description: ch.Description}
	var eps []outEpisode
	for _, item := range ch.Items {
		if item.Title == "" || item.Enclosure.URL == "" {
			continue
		}
		id := item.GUID
		if id == "" {
			id = item.Enclosure.URL
		}
		eps = append(eps, outEpisode{
			ID: id, Title: item.Title, Description: item.Description,
			PubDate: item.PubDate, AudioURL: item.Enclosure.URL, Duration: item.Duration,
		})
	}
	return p, eps, nil
}

func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
