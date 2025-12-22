//go:build darwin
// +build darwin

package screenshot

import (
	"errors"
)

// WindowInfo holds information about a window
type WindowInfo struct {
	Handle uintptr
	Title  string
}

// ListWindowsContaining returns all windows whose titles contain the given substring
// This is a stub for macOS - not implemented
func ListWindowsContaining(substring string) ([]WindowInfo, error) {
	return nil, errors.New("ListWindowsContaining not implemented on macOS")
}
