//go:build darwin
// +build darwin

package screenshot

import (
	"errors"
	"image"
)

// FindWindowWithPartialTitle is a stub for macOS
func FindWindowWithPartialTitle(titlePrefix string) (uintptr, error) {
	return 0, errors.New("screenshot functionality not implemented on macOS")
}

// CaptureWindow is a stub for macOS
func CaptureWindow(windowTitle string) (image.Image, error) {
	return nil, errors.New("screenshot functionality not implemented on macOS")
}

// TakeScreenshot is a stub for macOS
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, error) {
	return "", errors.New("screenshot functionality not implemented on macOS")
}
