//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package screenshot

import (
	"errors"
	"image"
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
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, error) {
	return "", errors.New("screenshot functionality not implemented on this platform")
}
