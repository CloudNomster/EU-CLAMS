//go:build linux
// +build linux

package screenshot

import (
	"errors"
	"image"
	"strings"
)

// FindWindowWithPartialTitle is a stub for Linux
func FindWindowWithPartialTitle(titlePrefix string) (uintptr, error) {
	return 0, errors.New("screenshot functionality not implemented on Linux")
}

// CaptureWindow is a stub for Linux
func CaptureWindow(windowTitle string) (image.Image, error) {
	return nil, errors.New("screenshot functionality not implemented on Linux")
}

// TakeScreenshot is a stub for Linux
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, string, error) {
	return "", "", errors.New("screenshot functionality not implemented on Linux")
}

// ExtractLocationFromWindowTitle attempts to extract a location name from the window title
// Location is expected to be in parentheses at the end of the title
func ExtractLocationFromWindowTitle(windowTitle string) string {
	// Check if the title has any content in parentheses at the end
	idx := strings.LastIndex(windowTitle, "(")
	if idx == -1 {
		return "" // No parentheses found
	}

	closingIdx := strings.LastIndex(windowTitle, ")")
	if closingIdx == -1 || closingIdx < idx {
		return "" // No closing parenthesis or it's before the opening one
	}

	// Extract the content between the parentheses
	locationName := windowTitle[idx+1 : closingIdx]
	return strings.TrimSpace(locationName)
}
