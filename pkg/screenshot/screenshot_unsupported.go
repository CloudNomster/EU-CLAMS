//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package screenshot

import (
	"errors"
	"image"
	"strings"
)

// FindWindowWithPartialTitle is a stub for unsupported platforms
func FindWindowWithPartialTitle(titlePrefix string) (uintptr, error) {
	return 0, errors.New("screenshot functionality not implemented on this platform")
}

// CaptureWindow is a stub for unsupported platforms
func CaptureWindow(windowTitle string) (image.Image, error) {
	return nil, errors.New("screenshot functionality not implemented on this platform")
}

// TakeScreenshot is a stub for unsupported platforms
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, string, error) {
	return "", "", errors.New("screenshot functionality not implemented on this platform")
}

// ExtractLocationFromWindowTitle attempts to extract a location name from the window title
// Location is expected to be in square brackets [] in the window title
// e.g., "Entropia Universe Client (64 bit) [Calypso]"
func ExtractLocationFromWindowTitle(windowTitle string) string {
	// Check if the title has any content in square brackets
	idx := strings.LastIndex(windowTitle, "[")
	if idx == -1 {
		return "" // No square brackets found
	}

	closingIdx := strings.LastIndex(windowTitle, "]")
	if closingIdx == -1 || closingIdx < idx {
		return "" // No closing bracket or it's before the opening one
	}

	// Extract the content between the square brackets
	locationName := windowTitle[idx+1 : closingIdx]
	return strings.TrimSpace(locationName)
}
