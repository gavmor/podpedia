package types

// Episode represents the raw data ingested from the RSS feed
type Episode struct {
	ID           string
	Title        string
	AudioURL     string
	Description  string
	Transcript   string
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
