package types

// Podcast represents the high-level metadata of a podcast feed
type Podcast struct {
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	URL         string   `xml:"-"`
	Author      string   `xml:"-"` // From iTunes/DublinCore
	Categories  []string `xml:"-"` // From iTunes
}

// Episode represents an individual podcast episode from the RSS feed
type Episode struct {
	ID          string   `xml:"guid" json:"id"`
	Title       string   `xml:"title" json:"title"`
	AudioURL    string   `xml:"-" json:"audio_url"`
	PubDate     string   `xml:"pubDate" json:"pub_date"`
	Description string   `xml:"description" json:"description"`
	Author      string   `xml:"-" json:"author"`
	Duration    string   `xml:"-" json:"duration"`
	Explicit    bool     `xml:"-" json:"explicit"`
	Categories  []string `xml:"-" json:"categories"`
	Transcript  string   `xml:"-" json:"transcript"`
}

// Guest represents an individual extracted by the LLM
type Guest struct {
	Name       string `json:"name"`
	Background string `json:"background"`
	Ideology   string `json:"ideology"`
}

// Company represents an institution extracted by the LLM
type Company struct {
	Name          string `json:"name"`
	BusinessModel string `json:"business_model"`
	Customers     string `json:"customers"`
}

// EncyclopediaEntry is the final structured output JSON schema
type EncyclopediaEntry struct {
	EpisodeID string    `json:"episode_id"`
	Guests    []Guest   `json:"guests"`
	Companies []Company `json:"companies"`
}
