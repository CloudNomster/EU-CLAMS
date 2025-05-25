package model

// GlobalEntryJSON is a JSON serialization-friendly version of GlobalEntry
type GlobalEntryJSON struct {
Timestamp  string  `json:"timestamp"`
Type       string  `json:"type"`
PlayerName string  `json:"player"`
TeamName   string  `json:"team,omitempty"`
Target     string  `json:"target"`
Value      float64 `json:"value"`
Location   string  `json:"location,omitempty"`
IsHof      bool    `json:"is_hof"`
RawMessage string  `json:"raw_message,omitempty"`
}
