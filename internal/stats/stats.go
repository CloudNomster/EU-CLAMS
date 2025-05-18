package stats

import (
	"eu-clams/internal/storage"
	"fmt"
	"sort"
	"strings"
	"time"
)

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

// GenerateStats generates statistics from database entries
func GenerateStats(db *storage.EntropyDB) Stats {
	stats := Stats{
		ByType:     make(map[string]int),
		ByLocation: make(map[string]int),
	}

	globals := db.GetPlayerGlobals()
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

// FormatStatsReport formats statistics into a readable report
func FormatStatsReport(stats Stats, playerName string, teamName string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Statistics for player: %s\n", playerName))
	if teamName != "" {
		b.WriteString(fmt.Sprintf("Team: %s\n", teamName))
	}
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	b.WriteString(fmt.Sprintf("Total globals: %d\n", stats.TotalGlobals))
	b.WriteString(fmt.Sprintf("Total HoFs: %d\n", stats.TotalHofs))
	b.WriteString(fmt.Sprintf("Total PED value: %.2f\n", stats.TotalValue))

	if stats.HighestValue > 0 {
		b.WriteString(fmt.Sprintf("Highest value: %.2f PED (%s)\n\n", stats.HighestValue, stats.HighestValueItem))
	}

	if len(stats.ByType) > 0 {
		b.WriteString("Globals by type:\n")
		for typ, count := range stats.ByType {
			b.WriteString(fmt.Sprintf("  %s: %d\n", strings.Title(typ), count))
		}
		b.WriteString("\n")
	}

	if len(stats.ByLocation) > 0 {
		b.WriteString("Globals by location:\n")

		// Sort locations by count
		type locCount struct {
			Location string
			Count    int
		}

		var locations []locCount
		for loc, count := range stats.ByLocation {
			locations = append(locations, locCount{loc, count})
		}

		sort.Slice(locations, func(i, j int) bool {
			return locations[i].Count > locations[j].Count
		})

		for _, lc := range locations {
			b.WriteString(fmt.Sprintf("  %s: %d\n", lc.Location, lc.Count))
		}
	}

	return b.String()
}
