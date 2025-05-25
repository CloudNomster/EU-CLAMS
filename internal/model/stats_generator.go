package model

import (
	"strings"
)

// GenerateStatsFromGlobals generates statistics based on a slice of GlobalEntry
func GenerateStatsFromGlobals(globals []GlobalEntry) Stats {
	stats := Stats{
		ByType:     make(map[string]int),
		ByLocation: make(map[string]int),
	}

	stats.TotalGlobals = len(globals)

	for _, entry := range globals {
		// Update total value
		stats.TotalValue += entry.Value

		// Track HoFs
		if entry.IsHof {
			stats.TotalHofs++
		}

		// Track highest value
		if entry.Value > stats.HighestValue {
			stats.HighestValue = entry.Value
			stats.HighestValueItem = entry.Target
		}

		// Count by type
		stats.ByType[entry.Type]++

		// Count by location
		if entry.Location != "" {
			location := strings.TrimSuffix(entry.Location, "!")
			stats.ByLocation[location]++
		}
	}

	return stats
}
