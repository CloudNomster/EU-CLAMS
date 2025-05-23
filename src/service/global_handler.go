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
				// Wait a short time to allow the UI to update before taking the screenshot
				time.Sleep(500 * time.Millisecond)

				screenshotPath, err := screenshotMgr.TakeScreenshotForGlobal(&e)
				if err != nil {
					s.log.Error("Failed to take screenshot: %v", err)
				} else {
					s.log.Info("Screenshot taken: %s", screenshotPath)
				}
			}(entry)
		}
	}
}
