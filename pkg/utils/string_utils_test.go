package utils

import "testing"

func TestTrimAndLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Should trim and lowercase",
			input:    "  Sample Text  ",
			expected: "sample text",
		},
		{
			name:     "Already lowercase and trimmed",
			input:    "test",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimAndLower(tt.input)
			if result != tt.expected {
				t.Errorf("TrimAndLower() = %v, want %v", result, tt.expected)
			}
		})
	}
}
