package service

import (
	"eu-clams/internal/storage"
	"os"
	"path/filepath"
	"time"
)

// GlobalHandler defines a function that processes a new global entry
type GlobalHandler func(entry *storage.GlobalEntry) error

// HandleNewGlobals processes new globals detected in the chat log
func (s *DataProcessorService) HandleNewGlobals(newEntries []storage.GlobalEntry) {
	if len(newEntries) == 0 {
		return
	}

	// Create screenshot manager if screenshots are enabled and not in import mode
	var screenshotMgr *ScreenshotManager
	if s.config.EnableScreenshots && !s.isImportMode {
		s.log.Info("Screenshot capturing is enabled for new globals")
	} else if s.isImportMode {
		s.log.Info("Screenshot capturing is disabled during import")
	}

	if s.config.EnableScreenshots && !s.isImportMode {
		// Ensure screenshot directory is absolute
		screenshotDir := s.config.ScreenshotDirectory
		if !filepath.IsAbs(screenshotDir) {
			// Use executable directory as base
			exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err == nil {
				screenshotDir = filepath.Join(exeDir, screenshotDir)
			}
		}
		// Create screenshot directory if it doesn't exist
		if err := os.MkdirAll(screenshotDir, 0755); err != nil {
			s.log.Error("Failed to create screenshot directory: %v", err)
		} else {
			screenshotMgr = NewScreenshotManager(
				screenshotDir,
				s.config.GameWindowTitle,
				s.config.EnableScreenshots,
				s.config.ScreenshotDelay,
			)
		}
	}

	for _, entry := range newEntries {
		// Take screenshot for player's globals or HOFs
		isRelevantToPlayer := s.config.PlayerName == "" ||
			entry.PlayerName == s.config.PlayerName
		isRelevantToTeam := s.config.TeamName == "" ||
			entry.TeamName == s.config.TeamName
		// Only take screenshots for player's/team's globals or any HOFs
		if screenshotMgr != nil && (entry.IsHof || isRelevantToPlayer || isRelevantToTeam) {
			go func(e storage.GlobalEntry) {
				// Wait for the configured amount of time to allow the UI to update before taking the screenshot
				time.Sleep(screenshotMgr.captureDelay)

				screenshotPath, err := screenshotMgr.TakeScreenshotForGlobal(&e)
				if err != nil {
					s.log.Error("Failed to take screenshot: %v", err)
				} else {
					s.log.Info("Screenshot taken: %s", screenshotPath) // Check if location was extracted from window title - if so, update the DB
					if e.Location != entry.Location && e.Location != "" {
						s.log.Info("Extracted location '%s' from window title for global: %s",
							e.Location, e.RawMessage)

						// Find and update this entry in the database
						s.db.UpdateGlobalLocation(&e)

						// Save the database to ensure the location is persisted
						dbPath := s.config.DatabasePath
						if !filepath.IsAbs(dbPath) {
							dbPath = filepath.Join(filepath.Dir(os.Args[0]), dbPath)
						}
						if err := s.db.SaveDatabase(dbPath, s.log); err != nil {
							s.log.Error("Failed to save database after location update: %v", err)
						} else {
							s.log.Info("Database updated with location: %s", e.Location)
						}

						// Broadcast updated stats since locations have changed
						statsData := s.db.GetStatsData()
						BroadcastToWebServices("stats_update", statsData)
					}
				}
			}(entry)
		}

		// Broadcast event to web services
		if entry.IsHof {
			// Broadcast as a HoF entry
			BroadcastToWebServices("new_hof", entry)
		} else {
			// Broadcast as a global entry
			BroadcastToWebServices("new_global", entry)
		}

		// Also broadcast updated stats
		statsData := s.db.GetStatsData()
		BroadcastToWebServices("stats_update", statsData)
	}
}
