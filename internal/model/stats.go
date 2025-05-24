package model

// Stats holds statistical information about globals and HoFs
type Stats struct {
	TotalGlobals     int
	TotalHofs        int
	HighestValue     float64
	HighestValueItem string
	TotalValue       float64
	ByType           map[string]int
	ByLocation       map[string]int
}

// GlobalEntry represents a single global message (copied for model independence)
type GlobalEntry struct {
	Timestamp  string  `json:"timestamp"`
	Type       string  `json:"type"`
	PlayerName string  `json:"playerName"`
	TeamName   string  `json:"teamName,omitempty"`
	Target     string  `json:"target"`
	Value      float64 `json:"value"`
	Location   string  `json:"location,omitempty"`
	IsHof      bool    `json:"isHof"`
}
