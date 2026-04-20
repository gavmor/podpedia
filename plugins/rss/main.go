//go:build wasip1

// Plugin: rss
// Receives raw RSS XML from the host (which fetched the feed URL) and returns
// a structured list of podcast episodes as JSON. No host imports needed —
// this plugin is pure computation.
package main

import (
	"encoding/json"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/mmcdole/gofeed"
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
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(raw)
	if err != nil {
		return outPodcast{}, nil, err
	}
	p := outPodcast{Title: feed.Title, Description: feed.Description}
	var eps []outEpisode
	for _, item := range feed.Items {
		if item.Title == "" || len(item.Enclosures) == 0 {
			continue
		}
		id := item.GUID
		if id == "" {
			id = item.Enclosures[0].URL
		}
		pubDate := ""
		if item.PublishedParsed != nil {
			pubDate = item.PublishedParsed.Format("Mon, 02 Jan 2006 15:04:05 +0000")
		} else if item.Published != "" {
			pubDate = item.Published
		}
		duration := ""
		if item.ITunesExt != nil {
			duration = item.ITunesExt.Duration
		}
		eps = append(eps, outEpisode{
			ID:          id,
			Title:       item.Title,
			Description: strings.TrimSpace(item.Description),
			PubDate:     pubDate,
			AudioURL:    item.Enclosures[0].URL,
			Duration:    duration,
		})
	}
	return p, eps, nil
}

func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
