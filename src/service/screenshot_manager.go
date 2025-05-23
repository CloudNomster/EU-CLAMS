package service

import (
	"eu-clams/internal/storage"
	"eu-clams/pkg/screenshot"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScreenshotManager handles taking screenshots for global and HoF events
type ScreenshotManager struct {
	screenshotDir   string
	gameWindowTitle string
	enabled         bool
	lastScreenshot  time.Time
}

// NewScreenshotManager creates a new screenshot manager
func NewScreenshotManager(screenshotDir, gameWindowTitle string, enabled bool) *ScreenshotManager {
	return &ScreenshotManager{
		screenshotDir:   screenshotDir,
		gameWindowTitle: gameWindowTitle,
		enabled:         enabled,
		lastScreenshot:  time.Time{},
	}
}

// TakeScreenshotForGlobal takes a screenshot for a global entry if enabled
func (sm *ScreenshotManager) TakeScreenshotForGlobal(entry *storage.GlobalEntry) (string, error) {
	if !sm.enabled {
		return "", fmt.Errorf("screenshots are disabled")
	}

	// Avoid taking multiple screenshots too quickly (within 2 seconds)
	if !sm.lastScreenshot.IsZero() && time.Since(sm.lastScreenshot) < 2*time.Second {
		return "", fmt.Errorf("screenshot already taken recently")
	}

	// Create prefix based on global type and whether it's a HoF
	prefix := entry.Type
	if entry.IsHof {
		prefix = "hof_" + prefix
	} else {
		prefix = "global_" + prefix
	}

	// Add player or team name to prefix
	if entry.TeamName != "" {
		prefix += "_" + strings.ReplaceAll(entry.TeamName, " ", "_")
	} else {
		prefix += "_" + strings.ReplaceAll(entry.PlayerName, " ", "_")
	}

	// Create absolute path for screenshot directory
	absScreenshotDir := sm.screenshotDir
	if !filepath.IsAbs(absScreenshotDir) {
		// Use executable directory as base
		exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err == nil {
			absScreenshotDir = filepath.Join(exeDir, sm.screenshotDir)
		}
	}

	// Take the screenshot
	filePath, err := screenshot.TakeScreenshot(sm.gameWindowTitle, absScreenshotDir, prefix)
	if err == nil {
		sm.lastScreenshot = time.Now()
	}

	return filePath, err
}

// IsEnabled returns whether screenshots are enabled
func (sm *ScreenshotManager) IsEnabled() bool {
	return sm.enabled
}

// SetEnabled enables or disables screenshots
func (sm *ScreenshotManager) SetEnabled(enabled bool) {
	sm.enabled = enabled
}
