package main

import (
	"eu-clams/pkg/screenshot"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Parse command line arguments
	windowTitle := "Entropia Universe Client (64 bit) [Calypso]"
	if len(os.Args) > 1 {
		windowTitle = os.Args[1]
	}

	fmt.Printf("Attempting to capture window titled: %s\n", windowTitle)

	// Create output directory
	outputDir := filepath.Join(".", "data", "screenshots", "test")
	os.MkdirAll(outputDir, 0755)

	// Take the screenshot
	filePath, err := screenshot.TakeScreenshot(windowTitle, outputDir, "test_capture")
	if err != nil {
		fmt.Printf("Error capturing screenshot: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Screenshot saved successfully: %s\n", filePath)
}
