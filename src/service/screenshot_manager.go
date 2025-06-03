package service

import (
	"eu-clams/internal/storage"
	"eu-clams/pkg/screenshot"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ScreenshotManager handles taking screenshots for global and HoF events
type ScreenshotManager struct {
	screenshotDir   string
	gameWindowTitle string
	enabled         bool
	lastScreenshot  time.Time
	platformSupport bool          // Whether the current platform supports screenshots
	captureDelay    time.Duration // Delay before taking screenshot to allow UI to update
}

// NewScreenshotManager creates a new screenshot manager
func NewScreenshotManager(screenshotDir, gameWindowTitle string, enabled bool, delaySeconds float64) *ScreenshotManager {
	// Check if we're on Windows, which is the only platform with screenshot support
	platformSupport := runtime.GOOS == "windows"

	return &ScreenshotManager{
		screenshotDir:   screenshotDir,
		gameWindowTitle: gameWindowTitle,
		enabled:         enabled && platformSupport, // Only enable if platform supports it
		lastScreenshot:  time.Time{},
		platformSupport: platformSupport,
		captureDelay:    time.Duration(delaySeconds * float64(time.Second)),
	}
}

// TakeScreenshotForGlobal takes a screenshot for a global entry if enabled
func (sm *ScreenshotManager) TakeScreenshotForGlobal(entry *storage.GlobalEntry) (string, error) {
	if !sm.enabled {
		if !sm.platformSupport {
			return "", fmt.Errorf("screenshots are not supported on %s platform", runtime.GOOS)
		}
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
	// Only allow enabling if platform supports it
	sm.enabled = enabled && sm.platformSupport
}

// IsPlatformSupported returns whether the current platform supports screenshots
func (sm *ScreenshotManager) IsPlatformSupported() bool {
	return sm.platformSupport
}
