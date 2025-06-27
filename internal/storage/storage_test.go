package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChatLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		line      string
		wantType  string
		wantHof   bool
		wantTeam  string
		wantValue float64
		wantError bool
	}{
		{
			name:      "Regular team kill",
			line:      "2025-05-16 10:00:00 [Globals] [] Team \"Test Team\" killed a creature (Test Beast) with a value of 100 PED",
			wantType:  "kill",
			wantHof:   false,
			wantTeam:  "Test Team",
			wantValue: 100,
		}, {
			name:      "Regular team kill with exclamation",
			line:      "2025-05-16 10:00:00 [Globals] [] Team \"Test Team\" killed a creature (Test Beast) with a value of 100 PED!",
			wantType:  "kill",
			wantHof:   false,
			wantTeam:  "Test Team",
			wantValue: 100,
		}, {
			name:      "HoF team kill with Hall of Fame text",
			line:      "2025-05-16 10:00:00 [Globals] [] Team \"Test Team\" killed a creature (Test Beast) with a value of 100 PED! A record has been added to the Hall of Fame!",
			wantType:  "kill",
			wantHof:   true,
			wantTeam:  "Test Team",
			wantValue: 100,
		},
		{
			name:      "Team kill with HTML entities in team name",
			line:      "2025-06-27 18:09:18 [Globals] [] Team &quot;***DeagleTeam***&quot; killed a creature (Eomon Old Alpha) with a value of 268 PED at OLA#63!",
			wantType:  "kill",
			wantHof:   false,
			wantTeam:  "***DeagleTeam***",
			wantValue: 268,
		},
		{
			name:      "Regular player kill",
			line:      "2025-05-16 10:01:00 [Globals] [] Test Player killed a creature (Test Beast) with a value of 50 PED",
			wantType:  "kill",
			wantHof:   false,
			wantValue: 50,
		}, {
			name:      "Regular player kill with exclamation",
			line:      "2025-05-16 10:01:00 [Globals] [] Test Player killed a creature (Test Beast) with a value of 50 PED!",
			wantType:  "kill",
			wantHof:   false,
			wantValue: 50,
		},
		{
			name:      "Regular team find",
			line:      "2025-05-16 10:02:00 [Globals] [] Team \"Test Team\" found a deposit (Test Material) with a value of 75 PED",
			wantType:  "find",
			wantHof:   false,
			wantTeam:  "Test Team",
			wantValue: 75,
		}, {
			name:      "Regular team find with exclamation",
			line:      "2025-05-16 10:02:00 [Globals] [] Team \"Test Team\" found a deposit (Test Material) with a value of 75 PED!",
			wantType:  "find",
			wantHof:   false,
			wantTeam:  "Test Team",
			wantValue: 75,
		},
		{
			name:      "Player craft",
			line:      "2025-05-16 10:05:00 [Globals] [] Test Player constructed an item (Test Item) worth 200 PED",
			wantType:  "craft",
			wantHof:   false,
			wantValue: 200,
		}, {
			name:      "Regular player craft with exclamation",
			line:      "2025-05-16 10:05:00 [Globals] [] Test Player constructed an item (Test Item) worth 200 PED!",
			wantType:  "craft",
			wantHof:   false,
			wantValue: 200,
		},
		{
			name:      "HoF player craft with Hall of Fame text",
			line:      "2025-05-16 10:05:00 [Globals] [] Test Player constructed an item (Test Item) worth 200 PED! A record has been added to the Hall of Fame!",
			wantType:  "craft",
			wantHof:   true,
			wantValue: 200,
		},
		{
			name:      "Invalid timestamp",
			line:      "Invalid-Date [Globals] [] Test Player killed a creature (Test Beast) with a value of 50 PED",
			wantError: true,
		},
		{
			name:      "Not a global message",
			line:      "2025-05-16 10:01:00 [Chat] Test Player: Hello world",
			wantError: false, // Not an error, just returns nil
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable for parallel execution
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entry, err := ParseChatLine(tt.line)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.wantError && entry == nil {
				// If we don't expect an error but don't expect a result either (like for non-global messages)
				if strings.Contains(tt.line, "[Globals]") {
					t.Errorf("Expected entry for global message, got nil")
				}
				return
			}

			if entry != nil {
				if entry.Type != tt.wantType {
					t.Errorf("Expected type %q, got %q", tt.wantType, entry.Type)
				}

				if entry.IsHof != tt.wantHof {
					t.Errorf("Expected IsHof=%v, got %v", tt.wantHof, entry.IsHof)
				}

				if tt.wantTeam != "" && entry.TeamName != tt.wantTeam {
					t.Errorf("Expected team %q, got %q", tt.wantTeam, entry.TeamName)
				}

				if entry.Value != tt.wantValue {
					t.Errorf("Expected value %.2f, got %.2f", tt.wantValue, entry.Value)
				}
			}
		})
	}
}

func TestGetHofEntries(t *testing.T) {
	t.Parallel()
	logContent := "2025-05-16 10:00:00 [Globals] [] Team \"Test Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
		"2025-05-16 10:01:00 [Globals] [] Test Player killed a creature (Test Beast) with a value of 50 PED!\n" +
		"2025-05-16 10:02:00 [Globals] [] Team \"Test Team\" found a deposit (Test Material) with a value of 75 PED\n" +
		"2025-05-16 10:05:00 [Globals] [] Test Player constructed an item (Test Item) worth 200 PED! A record has been added to the Hall of Fame!\n"

	// Create a temporary test file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "hof_test.log")

	if err := os.WriteFile(tmpFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	db := NewEntropyDB("", "")
	_, err := db.ProcessChatLog(tmpFile, nil, nil)
	if err != nil {
		t.Errorf("ProcessChatLog failed: %v", err)
		return
	}
	// Get HoF entries
	hofEntries := db.GetHofEntries()

	// Should have exactly 1 HoF entry
	if len(hofEntries) != 1 {
		t.Errorf("Expected 1 HoF entry, got %d", len(hofEntries))
	}
	// Verify the HoF entries are the correct ones
	for _, entry := range hofEntries {
		if !entry.IsHof {
			t.Errorf("GetHofEntries returned a non-HoF entry: %v", entry)
			continue
		}

		// The HoF entry should be the craft with value 200
		if entry.Type == "craft" && entry.Value != 200 {
			t.Errorf("Expected HoF craft entry with value 200, got %.2f", entry.Value)
		}
	}
}

func TestRealWorldHofParsing(t *testing.T) {
	t.Parallel()

	// Test with the actual example from our test log
	hofLine := "2025-05-06 15:15:45 [Globals] [] Test Player killed a creature (Lairkeeper, Brood of Unruly) with a value of 119 PED! A record has been added to the Hall of Fame!"

	entry, err := ParseChatLine(hofLine)
	if err != nil {
		t.Errorf("Failed to parse real HoF line: %v", err)
		return
	}

	if entry == nil {
		t.Errorf("ParseChatLine returned nil for a valid HoF line")
		return
	}

	if !entry.IsHof {
		t.Errorf("Expected IsHof=true for line with 'Hall of Fame' text, got false")
	}

	if entry.Type != "kill" {
		t.Errorf("Expected type 'kill', got %q", entry.Type)
	}

	if entry.Value != 119 {
		t.Errorf("Expected value 119, got %.2f", entry.Value)
	}

	if entry.PlayerName != "Test Player" {
		t.Errorf("Expected player 'Test Player', got %q", entry.PlayerName)
	}

	if entry.Target != "Lairkeeper, Brood of Unruly" {
		t.Errorf("Expected target 'Lairkeeper, Brood of Unruly', got %q", entry.Target)
	}
}

func TestTeamNameFiltering(t *testing.T) {
	t.Parallel() // Allow test to run in parallel

	tests := []struct {
		name        string
		playerName  string
		teamName    string
		logContent  string
		wantCount   int
		wantPlayers []string
		wantTeams   []string
	}{
		{
			name:     "Team filter only",
			teamName: "Alpha Team",
			logContent: "2025-05-16 10:00:00 [Globals] [] Team \"Alpha Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
				"2025-05-16 10:01:00 [Globals] [] John Doe killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount: 1,
			wantTeams: []string{"Alpha Team"},
		},
		{
			name:       "Player filter only",
			playerName: "John Doe",
			logContent: "2025-05-16 10:00:00 [Globals] [] Team \"Alpha Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
				"2025-05-16 10:01:00 [Globals] [] John Doe killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount:   1,
			wantPlayers: []string{"John Doe"},
		},
		{
			name:       "Both team and player filter",
			playerName: "John Doe",
			teamName:   "Alpha Team",
			logContent: "2025-05-16 10:00:00 [Globals] [] Team \"Alpha Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
				"2025-05-16 10:01:00 [Globals] [] John Doe killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount:   2,
			wantPlayers: []string{"John Doe"},
			wantTeams:   []string{"Alpha Team"},
		},
		{
			name:     "Team filter with HTML entities - config without quotes",
			teamName: "***DeagleTeam***",
			logContent: "2025-06-27 18:09:18 [Globals] [] Team &quot;***DeagleTeam***&quot; killed a creature (Eomon Old Alpha) with a value of 268 PED at OLA#63!\n" +
				"2025-06-27 18:10:00 [Globals] [] Other Player killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount: 1,
			wantTeams: []string{"***DeagleTeam***"},
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable for parallel execution
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			// Create a temporary test file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.log")

			if err := os.WriteFile(tmpFile, []byte(tt.logContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			db := NewEntropyDB(tt.playerName, tt.teamName)
			count, err := db.ProcessChatLog(tmpFile, nil, nil)
			if err != nil {
				t.Errorf("ProcessChatLog failed: %v", err)
				return
			}

			if count != tt.wantCount {
				t.Errorf("Expected %d entries, got %d", tt.wantCount, count)
			}

			// Verify players and teams in results
			for _, wantPlayer := range tt.wantPlayers {
				found := false
				for _, entry := range db.Globals {
					if strings.EqualFold(entry.PlayerName, wantPlayer) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find player %q in results", wantPlayer)
				}
			}

			for _, wantTeam := range tt.wantTeams {
				found := false
				for _, entry := range db.Globals {
					if strings.EqualFold(entry.TeamName, wantTeam) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find team %q in results", wantTeam)
				}
			}
		})
	}
}

func TestTeamNameMatching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configTeam string
		entryTeam  string
		wantMatch  bool
	}{
		{
			name:       "Exact match without quotes",
			configTeam: "DeagleTeam",
			entryTeam:  "DeagleTeam",
			wantMatch:  true,
		},
		{
			name:       "Config has no quotes, entry has quotes",
			configTeam: "DeagleTeam",
			entryTeam:  "\"DeagleTeam\"",
			wantMatch:  true,
		},
		{
			name:       "Config has quotes, entry has no quotes",
			configTeam: "\"DeagleTeam\"",
			entryTeam:  "DeagleTeam",
			wantMatch:  true,
		},
		{
			name:       "Both have quotes",
			configTeam: "\"DeagleTeam\"",
			entryTeam:  "\"DeagleTeam\"",
			wantMatch:  true,
		},
		{
			name:       "Case insensitive match",
			configTeam: "deagleteam",
			entryTeam:  "DeagleTeam",
			wantMatch:  true,
		},
		{
			name:       "With special characters and HTML entities",
			configTeam: "***DeagleTeam***",
			entryTeam:  "&quot;***DeagleTeam***&quot;",
			wantMatch:  true,
		},
		{
			name:       "Different teams",
			configTeam: "TeamA",
			entryTeam:  "TeamB",
			wantMatch:  false,
		},
		{
			name:       "Empty config team",
			configTeam: "",
			entryTeam:  "DeagleTeam",
			wantMatch:  false,
		},
		{
			name:       "Empty entry team",
			configTeam: "DeagleTeam",
			entryTeam:  "",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := teamNamesMatch(tt.entryTeam, tt.configTeam)
			if result != tt.wantMatch {
				t.Errorf("teamNamesMatch(%q, %q) = %v, want %v",
					tt.entryTeam, tt.configTeam, result, tt.wantMatch)
			}
		})
	}
}

// TestProcessChatLogFromOffsetMethod tests the ProcessChatLogFromOffset method
func TestProcessChatLogFromOffsetMethod(t *testing.T) {
	t.Parallel() // Allow test to run in parallel

	tests := []struct {
		name        string
		playerName  string
		teamName    string
		logContent  string
		wantCount   int
		wantPlayers []string
		wantTeams   []string
	}{
		{
			name:     "Team filter only",
			teamName: "Alpha Team",
			logContent: "2025-05-16 10:00:00 [Globals] [] Team \"Alpha Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
				"2025-05-16 10:01:00 [Globals] [] John Doe killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount: 1,
			wantTeams: []string{"Alpha Team"},
		},
		{
			name:       "Player filter only",
			playerName: "John Doe",
			logContent: "2025-05-16 10:00:00 [Globals] [] Team \"Alpha Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
				"2025-05-16 10:01:00 [Globals] [] John Doe killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount:   1,
			wantPlayers: []string{"John Doe"},
		},
		{
			name:       "Both team and player filter",
			playerName: "John Doe",
			teamName:   "Alpha Team",
			logContent: "2025-05-16 10:00:00 [Globals] [] Team \"Alpha Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
				"2025-05-16 10:01:00 [Globals] [] John Doe killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount:   2,
			wantPlayers: []string{"John Doe"},
			wantTeams:   []string{"Alpha Team"},
		},
		{
			name:     "Team filter with HTML entities - config without quotes",
			teamName: "***DeagleTeam***",
			logContent: "2025-06-27 18:09:18 [Globals] [] Team &quot;***DeagleTeam***&quot; killed a creature (Eomon Old Alpha) with a value of 268 PED at OLA#63!\n" +
				"2025-06-27 18:10:00 [Globals] [] Other Player killed a creature (Test Beast) with a value of 50 PED\n",
			wantCount: 1,
			wantTeams: []string{"***DeagleTeam***"},
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable for parallel execution
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			// Create a temporary test file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.log")

			if err := os.WriteFile(tmpFile, []byte(tt.logContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			db := NewEntropyDB(tt.playerName, tt.teamName)
			count, err := db.ProcessChatLogFromOffset(tmpFile, 0, nil, nil)
			if err != nil {
				t.Errorf("ProcessChatLogFromOffset failed: %v", err)
				return
			}

			if count != tt.wantCount {
				t.Errorf("Expected %d entries, got %d", tt.wantCount, count)
			}

			// Verify players and teams in results
			for _, wantPlayer := range tt.wantPlayers {
				found := false
				for _, entry := range db.Globals {
					if strings.EqualFold(entry.PlayerName, wantPlayer) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find player %q in results", wantPlayer)
				}
			}

			for _, wantTeam := range tt.wantTeams {
				found := false
				for _, entry := range db.Globals {
					if strings.EqualFold(entry.TeamName, wantTeam) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find team %q in results", wantTeam)
				}
			}
		})
	}
}

// TestConsistentFiltering ensures that ProcessChatLog and ProcessChatLogFromOffset have identical filtering behavior
func TestConsistentFiltering(t *testing.T) {
	testData := "2025-05-16 10:00:00 [Globals] [] Team \"Test Team\" killed a creature (Test Beast) with a value of 100 PED\n" +
		"2025-05-16 10:01:00 [Globals] [] Test Player killed a creature (Test Beast) with a value of 50 PED\n" +
		"2025-05-16 10:02:00 [Globals] [] Random Player killed a creature (Test Beast) with a value of 75 PED\n"

	tests := []struct {
		name       string
		playerName string
		teamName   string
	}{
		{
			name:       "No filters",
			playerName: "",
			teamName:   "",
		},
		{
			name:       "Team filter only",
			playerName: "",
			teamName:   "Test Team",
		},
		{
			name:       "Player filter only",
			playerName: "Test Player",
			teamName:   "",
		},
		{
			name:       "Both filters",
			playerName: "Test Player",
			teamName:   "Test Team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary test file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.log")

			if err := os.WriteFile(tmpFile, []byte(testData), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Process with both methods
			db1 := NewEntropyDB(tt.playerName, tt.teamName)
			count1, err1 := db1.ProcessChatLog(tmpFile, nil, nil)
			if err1 != nil {
				t.Errorf("ProcessChatLog failed: %v", err1)
				return
			}

			db2 := NewEntropyDB(tt.playerName, tt.teamName)
			count2, err2 := db2.ProcessChatLogFromOffset(tmpFile, 0, nil, nil)
			if err2 != nil {
				t.Errorf("ProcessChatLogFromOffset failed: %v", err2)
				return
			}

			// Compare results
			if count1 != count2 {
				t.Errorf("Different counts: ProcessChatLog=%d, ProcessChatLogFromOffset=%d", count1, count2)
			}

			if len(db1.Globals) != len(db2.Globals) {
				t.Errorf("Different number of globals: ProcessChatLog=%d, ProcessChatLogFromOffset=%d",
					len(db1.Globals), len(db2.Globals))
			}
		})
	}
}
