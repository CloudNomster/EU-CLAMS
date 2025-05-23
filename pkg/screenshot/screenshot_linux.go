//go:build linux
// +build linux

package screenshot

import (
	"errors"
	"image"
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
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, error) {
	return "", errors.New("screenshot functionality not implemented on Linux")
}
