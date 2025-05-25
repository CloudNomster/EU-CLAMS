package stats

import (
	"eu-clams/internal/model"
	"fmt"
	"sort"
	"strings"
	"time"
)

// For backward compatibility
type Stats = model.Stats

// GenerateStats is a compatibility function that now delegates to the storage package
// It's kept for backwards compatibility
func GenerateStats(dbInterface interface{}) Stats {
	// For compatibility with existing code, we now expect the DB to provide its own stats
	if statsProvider, ok := dbInterface.(interface{ GetStatsData() model.Stats }); ok {
		return statsProvider.GetStatsData()
	}

	// Return empty stats if the interface doesn't match
	return Stats{
		ByType:     make(map[string]int),
		ByLocation: make(map[string]int),
	}
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
