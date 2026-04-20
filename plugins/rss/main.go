//go:build wasip1

// Plugin: rss
// Receives raw RSS XML from the host (which fetched the feed URL) and returns
// a structured list of podcast episodes as JSON. No host imports needed —
// this plugin is pure computation.
package main

import (
	"encoding/json"
	"strings"

	"github.com/gavmor/wasm-microkernel/guest-bindings/plugin_world"
	"github.com/mmcdole/gofeed"
)

func main() {}

func init() { plugin_world.SetExportsPluginWorld(&RSSPlugin{}) }

type RSSPlugin struct{}

func (r *RSSPlugin) Execute(reqJSON string) (plugin_world.Result[string, string], error) {
	var req struct {
		XML string `json:"xml"`
	}
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return plugin_world.Err[string, string]("bad request: " + err.Error()), nil
	}
	p, eps, err := parseRSS(req.XML)
	if err != nil {
		return plugin_world.Err[string, string]("rss parse: " + err.Error()), nil
	}
	out, _ := json.Marshal(map[string]any{"podcast": p, "episodes": eps})
	return plugin_world.Ok[string, string](string(out)), nil
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
