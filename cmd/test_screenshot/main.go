package main

import (
	"eu-clams/pkg/screenshot"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Get the current directory
	exeDir, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}

	baseDir := filepath.Dir(exeDir)
	screenshotDir := filepath.Join(baseDir, "data", "screenshots")

	// Ensure screenshot directory exists
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		fmt.Printf("Error creating screenshot directory: %v\n", err)
		return
	}

	fmt.Println("Taking screenshot of 'Entropia Universe Client' window...")

	// Generate a unique filename with timestamp
	// 	timestamp := time.Now().Format("2006-01-02_15-04-05")
	prefix := "test_screenshot"

	path, err := screenshot.TakeScreenshot("Entropia Universe Client (64 bit) [Calypso]", screenshotDir, prefix)
	if err != nil {
		fmt.Printf("Error taking screenshot: %v\n", err)
		return
	}

	fmt.Printf("Screenshot saved to: %s\n", path)
}
