package storage

import (
	"bufio"
	"eu-clams/internal/logger"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// GlobalEntry represents a single global message
type GlobalEntry struct {
	Timestamp  time.Time `yaml:"timestamp" json:"timestamp"`
	Type       string    `yaml:"type" json:"type"` // e.g., "kill", "craft", "find"
	PlayerName string    `yaml:"player" json:"player"`
	TeamName   string    `yaml:"team,omitempty" json:"team,omitempty"`
	Target     string    `yaml:"target" json:"target"` // creature/item name
	Value      float64   `yaml:"value" json:"value"`   // PED value
	Location   string    `yaml:"location,omitempty" json:"location,omitempty"`
	IsHof      bool      `yaml:"is_hof" json:"is_hof"`
	RawMessage string    `yaml:"raw_message" json:"raw_message"`
}

// EntropyDB is the main structure for storing EU data
type EntropyDB struct {
	Globals           []GlobalEntry `yaml:"globals"`
	PlayerName        string        `yaml:"player_name,omitempty"`
	TeamName          string        `yaml:"team_name,omitempty"`
	LastProcessed     time.Time     `yaml:"last_processed,omitempty"`
	LastProcessedSize int64         `yaml:"last_processed_size,omitempty"`
}

// NewEntropyDB creates a new empty database
func NewEntropyDB(playerName string, teamName string) *EntropyDB {
	return &EntropyDB{
		Globals:    []GlobalEntry{},
		PlayerName: playerName,
		TeamName:   teamName,
	}
}

// regular expressions for parsing different types of global messages
var ( // For team kills
	teamKillRegex = regexp.MustCompile(`\[\s*Globals\s*\]\s*\[\s*\]\s*Team\s*"([^"]+)"\s*killed\s*a\s*creature\s*\(([^)]+)\)\s*with\s*a\s*value\s*of\s*(\d+)\s*PED(?:\s*at\s*([^!]+))?(!)?(?:\s*A\s*record\s*has\s*been\s*added\s*to\s*the\s*Hall\s*of\s*Fame!)?`)

	// For individual kills
	playerKillRegex = regexp.MustCompile(`\[\s*Globals\s*\]\s*\[\s*\]\s*([^\s]+(?:\s+[^\s]+){0,3})\s*(?:as|has|have)?\s*killed\s*a\s*creature\s*\(([^)]+)\)\s*with\s*a\s*value\s*of\s*(\d+)\s*PED(?:\s*at\s*([^!]+))?(!)?(?:\s*A\s*record\s*has\s*been\s*added\s*to\s*the\s*Hall\s*of\s*Fame!)?`)

	// For crafting
	craftRegex = regexp.MustCompile(`\[\s*Globals\s*\]\s*\[\s*\]\s*([^\s]+(?:\s+[^\s]+){0,3})\s*constructed\s*an\s*item\s*\(([^)]+)\)\s*worth\s*(\d+)\s*PED(!)?(?:\s*A\s*record\s*has\s*been\s*added\s*to\s*the\s*Hall\s*of\s*Fame!)?`)

	// For mining/deposits
	findRegex = regexp.MustCompile(`\[\s*Globals\s*\]\s*\[\s*\]\s*(?:Team\s*"([^"]+)"|([^\s]+(?:\s+[^\s]+){0,3}))\s*found\s*a\s*deposit\s*\(([^)]+)\)\s*with\s*a\s*value\s*of\s*(\d+)\s*PED(!)?(?:\s*A\s*record\s*has\s*been\s*added\s*to\s*the\s*Hall\s*of\s*Fame!)?`)
)

// ParseChatLine parses a single line from the chat log and returns a GlobalEntry if it's a global message
func ParseChatLine(line string) (*GlobalEntry, error) {
	// Skip if not a global message
	if !strings.Contains(line, "[Globals]") {
		return nil, nil
	}

	// Extract timestamp
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid line format: %s", line)
	}

	dateStr := parts[0] + " " + parts[1]
	timestamp, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %s", dateStr)
	}

	entry := GlobalEntry{
		Timestamp:  timestamp,
		RawMessage: line,
	}

	// Try to match against different global message patterns
	if matches := teamKillRegex.FindStringSubmatch(line); matches != nil {
		entry.Type = "kill"
		entry.TeamName = matches[1]
		entry.Target = matches[2]
		entry.Value = parseValue(matches[3])
		if len(matches) > 4 && matches[4] != "" {
			entry.Location = strings.TrimSpace(matches[4])
		}
		// Check for Hall of Fame indicator - only use the explicit "Hall of Fame" text
		entry.IsHof = strings.Contains(line, "Hall of Fame")
		return &entry, nil
	}

	if matches := playerKillRegex.FindStringSubmatch(line); matches != nil {
		entry.Type = "kill"
		entry.PlayerName = matches[1]
		entry.Target = matches[2]
		entry.Value = parseValue(matches[3])
		if len(matches) > 4 && matches[4] != "" {
			entry.Location = strings.TrimSpace(matches[4])
		}
		// Check for Hall of Fame indicator - only use the explicit "Hall of Fame" text
		entry.IsHof = strings.Contains(line, "Hall of Fame")
		return &entry, nil
	}

	if matches := craftRegex.FindStringSubmatch(line); matches != nil {
		entry.Type = "craft"
		entry.PlayerName = matches[1]
		entry.Target = matches[2]
		entry.Value = parseValue(matches[3])
		// Check for Hall of Fame indicator - only use the explicit "Hall of Fame" text
		entry.IsHof = strings.Contains(line, "Hall of Fame")
		return &entry, nil
	}

	if matches := findRegex.FindStringSubmatch(line); matches != nil {
		entry.Type = "find"
		if matches[1] != "" {
			entry.TeamName = matches[1]
		} else {
			entry.PlayerName = matches[2]
		}
		entry.Target = matches[3]
		entry.Value = parseValue(matches[4])
		// Check for Hall of Fame indicator - only use the explicit "Hall of Fame" text
		entry.IsHof = strings.Contains(line, "Hall of Fame")
		return &entry, nil
	}

	// Return nil if we couldn't parse this global message
	// It could be another type we're not handling yet
	return nil, nil
}

// parseValue converts a string PED value to float64
func parseValue(val string) float64 {
	var value float64
	fmt.Sscanf(val, "%f", &value)
	return value
}

// ProcessChatLogFromOffset reads a chat log file from a specific offset and extracts global messages
func (db *EntropyDB) ProcessChatLogFromOffset(logPath string, offset int64, progressChan chan<- float64, logger *logger.Logger) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}

	file, err := os.Open(logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open chat log: %w", err)
	}
	defer file.Close()

	// Get file size for progress tracking
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	totalSize := float64(fileInfo.Size())

	// Seek to the offset
	if _, err := file.Seek(offset, 0); err != nil {
		return 0, fmt.Errorf("failed to seek to offset: %w", err)
	}

	if logger != nil {
		logger.Debug("Processing chat log from offset %d (%.1f%%)", offset, (float64(offset)/totalSize)*100)
		logger.Debug("Player filter: %s, Team filter: %s", db.PlayerName, db.TeamName)
	}

	scanner := bufio.NewScanner(file)
	count := 0
	bytesRead := float64(offset)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		bytesRead += float64(len(line) + 1) // +1 for newline

		if progressChan != nil {
			select {
			case progressChan <- (bytesRead / totalSize) * 100:
				// Progress sent successfully
			default:
				// Channel is full or closed, skip progress update
			}
		}

		entry, err := ParseChatLine(line)
		if err != nil {
			if logger != nil {
				logger.Error("Error parsing line %d: %v\nLine content: %s", lineNum, err, line)
			}
			continue
		}

		if entry != nil {
			if logger != nil {
				logger.Info("Line %d - Found global: Type=%s, Player=%s, Team=%s, Target=%s, Value=%.2f",
					lineNum, entry.Type, entry.PlayerName, entry.TeamName, entry.Target, entry.Value)
			}

			matchPlayer := db.PlayerName == "" || strings.EqualFold(entry.PlayerName, db.PlayerName)
			matchTeam := db.TeamName == "" || strings.EqualFold(entry.TeamName, db.TeamName)

			// Determine if we should include this entry based on filter settings
			shouldInclude := false

			if db.PlayerName != "" && db.TeamName != "" {
				// Both filters active - include if either matches
				shouldInclude = matchPlayer || matchTeam
			} else if db.PlayerName != "" {
				// Only player filter active
				shouldInclude = matchPlayer && entry.PlayerName != "" // Must be a player entry
			} else if db.TeamName != "" {
				// Only team filter active
				shouldInclude = matchTeam && entry.TeamName != "" // Must be a team entry
			} else {
				// No filters active - include all entries
				shouldInclude = true
			}

			if shouldInclude {
				db.Globals = append(db.Globals, *entry)
				count++
				if logger != nil {
					logger.Info("Added global from line %d (total: %d)", lineNum, count)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("error reading chat log: %w", err)
	}

	if logger != nil && count > 0 {
		logger.Debug("Finished processing chat log from offset. Added %d new globals.", count)

		// Log all stored globals for debugging
		logger.Info("Stored globals (%d):", len(db.Globals))
		for i, g := range db.Globals {
			logger.Debug("%d. Type=%s, Player=%s, Team=%s, Target=%s, Value=%.2f",
				i+1, g.Type, g.PlayerName, g.TeamName, g.Target, g.Value)
		}
	}

	db.LastProcessedSize = fileInfo.Size()
	db.LastProcessed = time.Now()
	return count, nil
}

// ProcessChatLog reads a chat log file and extracts all global messages
// Only processes globals relevant to the player if playerName is specified
func (db *EntropyDB) ProcessChatLog(logPath string, progressChan chan<- float64, logger *logger.Logger) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("chat log file does not exist: %s", logPath)
	}

	file, err := os.Open(logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open chat log: %w", err)
	}
	defer file.Close()

	// Get file size for progress tracking
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	totalSize := float64(fileInfo.Size())

	if logger != nil {
		logger.Debug("Starting to process chat log: %s (size: %d)", logPath, fileInfo.Size())
		logger.Debug("Player filter: %s, Team filter: %s", db.PlayerName, db.TeamName)
	}

	scanner := bufio.NewScanner(file)
	count := 0
	bytesRead := float64(0)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		bytesRead += float64(len(line) + 1) // +1 for newline

		if progressChan != nil {
			select {
			case progressChan <- (bytesRead / totalSize) * 100:
				// Progress sent successfully
			default:
				// Channel is full or closed, skip progress update
			}
		}
		entry, err := ParseChatLine(line)
		if err != nil {
			if logger != nil {
				logger.Error("Error parsing line %d: %v\nLine content: %s", lineNum, err, line)
			}
			continue
		}

		if entry != nil {
			if logger != nil {
				logger.Debug("Line %d - Found global: Type=%s, Player=%s, Team=%s, Target=%s, Value=%.2f",
					lineNum, entry.Type, entry.PlayerName, entry.TeamName, entry.Target, entry.Value)
			} // Include the entry if any of these are true:
			// 1. No player/team filtering is enabled
			// 2. It's the player's own global
			// 3. It's from the player's team
			matchPlayer := db.PlayerName == "" || strings.EqualFold(entry.PlayerName, db.PlayerName)
			matchTeam := db.TeamName == "" || strings.EqualFold(entry.TeamName, db.TeamName)

			// Determine if we should include this entry based on filter settings
			shouldInclude := false

			if db.PlayerName != "" && db.TeamName != "" {
				// Both filters active - include if either matches
				shouldInclude = matchPlayer || matchTeam
			} else if db.PlayerName != "" {
				// Only player filter active
				shouldInclude = matchPlayer && entry.PlayerName != "" // Must be a player entry
			} else if db.TeamName != "" {
				// Only team filter active
				shouldInclude = matchTeam && entry.TeamName != "" // Must be a team entry
			} else {
				// No filters active - include all entries
				shouldInclude = true
			}

			if shouldInclude {
				db.Globals = append(db.Globals, *entry)
				count++
				if logger != nil {
					logger.Info("Added global from line %d (total: %d)", lineNum, count)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("error reading chat log: %w", err)
	}

	if logger != nil {
		logger.Debug("Finished processing chat log. Found %d globals.", count)

		// Log all stored globals for debugging
		logger.Debug("Stored globals (%d):", len(db.Globals))
		for i, g := range db.Globals {
			logger.Debug("%d. Type=%s, Player=%s, Team=%s, Target=%s, Value=%.2f",
				i+1, g.Type, g.PlayerName, g.TeamName, g.Target, g.Value)
		}
	}

	db.LastProcessed = time.Now()
	db.LastProcessedSize = fileInfo.Size()
	return count, nil
}

// GetHofEntries returns all HoF entries, ordered by timestamp (newest first)
func (db *EntropyDB) GetHofEntries() []GlobalEntry {
	var hofs []GlobalEntry
	for _, entry := range db.Globals {
		if entry.IsHof {
			hofs = append(hofs, entry)
		}
	}

	// Sort by timestamp in descending order (newest first)
	sort.Slice(hofs, func(i, j int) bool {
		return hofs[i].Timestamp.After(hofs[j].Timestamp)
	})

	return hofs
}

// GetEntriesByType returns all entries of a specific type
func (db *EntropyDB) GetEntriesByType(entryType string) []GlobalEntry {
	var results []GlobalEntry
	for _, entry := range db.Globals {
		if entry.Type == entryType {
			results = append(results, entry)
		}
	}
	return results
}

// GetEntriesByPlayer returns all entries for a specific player
func (db *EntropyDB) GetEntriesByPlayer(playerName string) []GlobalEntry {
	var results []GlobalEntry
	for _, entry := range db.Globals {
		if strings.Contains(strings.ToLower(entry.PlayerName), strings.ToLower(playerName)) {
			results = append(results, entry)
		}
	}
	return results
}

// GetEntriesByValue returns all entries with value >= minValue
func (db *EntropyDB) GetEntriesByValue(minValue float64) []GlobalEntry {
	var results []GlobalEntry
	for _, entry := range db.Globals {
		if entry.Value >= minValue {
			results = append(results, entry)
		}
	}
	return results
}
