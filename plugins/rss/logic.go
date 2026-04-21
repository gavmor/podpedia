package main

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

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
