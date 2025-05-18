package utils

import (
	"strings"
)

// TrimAndLower trims a string and converts it to lowercase
func TrimAndLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
