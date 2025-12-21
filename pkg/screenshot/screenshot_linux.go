//go:build linux
// +build linux

package screenshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SwayNode represents a node in the Sway tree structure
type SwayNode struct {
	ID               int64       `json:"id"`
	Name             string      `json:"name"`
	Type             string      `json:"type"`
	Nodes            []*SwayNode `json:"nodes"`
	FloatingNodes    []*SwayNode `json:"floating_nodes"`
	Rect             SwayRect    `json:"rect"`
	Visible          bool        `json:"visible"`
	Pid              int         `json:"pid"`
	AppID            string      `json:"app_id"`
	WindowProperties *WindowProp `json:"window_properties"`
	Focused          bool        `json:"focused"`
}

type SwayRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WindowProp struct {
	Class    string `json:"class"`
	Instance string `json:"instance"`
	Title    string `json:"title"`
}

// FindWindowWithPartialTitle finds a window where the title starts with the given prefix
// Returns the Sway node ID as uintptr (although it's just an ID)
func FindWindowWithPartialTitle(titlePrefix string) (uintptr, error) {
	node, err := findNodeByTitle(titlePrefix)
	if err != nil {
		return 0, err
	}
	return uintptr(node.ID), nil
}

// GetFullWindowTitle finds the window with the given title prefix and returns its full title
func GetFullWindowTitle(windowTitle string) (string, error) {
	node, err := findNodeByTitle(windowTitle)
	if err != nil {
		return "", err
	}
	return node.Name, nil
}

// IsWindowVisible checks if the window with the given handle (ID) is visible
// Note: In sway, we might check if it's on an active workspace, but simply returning true if found is often enough for "visible" checks in this context,
// or we can check the "visible" property if available in the tree (it is)
func IsWindowVisible(hwnd uintptr) bool {
	// Re-query the tree to check visibility? Or just assume if we found it it's "there".
	// The sway tree output has a "visible" boolean for workspaces/views.
	// We'll simplisticly check if we can find it again.
	// For now, let's assume if we can capture it, it's fine.
	// Windows implementation checks IsWindowVisible.
	// Let's implement a quick check.

	// We need to re-fetch the tree to get current state
	cmd := exec.Command("swaymsg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	var root SwayNode
	if err := json.Unmarshal(output, &root); err != nil {
		return false
	}

	found := findNodeByID(&root, int64(hwnd))
	return found != nil && found.Visible // visible property might be on parent workspace, but simple check for now
}

func findNodeByID(node *SwayNode, id int64) *SwayNode {
	if node.ID == id {
		return node
	}
	for _, child := range node.Nodes {
		if found := findNodeByID(child, id); found != nil {
			return found
		}
	}
	for _, child := range node.FloatingNodes {
		if found := findNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// findNodeByTitle uses swaymsg to find a window
func findNodeByTitle(titlePrefix string) (*SwayNode, error) {
	cmd := exec.Command("swaymsg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute swaymsg: %w", err)
	}

	var root SwayNode
	if err := json.Unmarshal(output, &root); err != nil {
		return nil, fmt.Errorf("failed to parse sway tree: %w", err)
	}

	found := searchTree(&root, strings.ToLower(titlePrefix))
	if found == nil {
		return nil, fmt.Errorf("window not found with prefix: %s", titlePrefix)
	}

	return found, nil
}

func searchTree(node *SwayNode, titlePrefixLower string) *SwayNode {
	// Check current node
	// Check Name (title)
	if node.Name != "" && strings.HasPrefix(strings.ToLower(node.Name), titlePrefixLower) {
		// Exclude some non-window nodes if strictly needed, but title match is usually good enough
		if node.Type == "con" || node.Type == "floating_con" {
			return node
		}
	}

	// Recurse children
	for _, child := range node.Nodes {
		if found := searchTree(child, titlePrefixLower); found != nil {
			return found
		}
	}
	for _, child := range node.FloatingNodes {
		if found := searchTree(child, titlePrefixLower); found != nil {
			return found
		}
	}

	return nil
}

// CaptureWindow captures a screenshot of the window with the given title
func CaptureWindow(windowTitle string) (image.Image, error) {
	node, err := findNodeByTitle(windowTitle)
	if err != nil {
		return nil, err
	}

	// Construct geometry string "X,Y WxH"
	geometry := fmt.Sprintf("%d,%d %dx%d", node.Rect.X, node.Rect.Y, node.Rect.Width, node.Rect.Height)

	// Use grim to capture
	cmd := exec.Command("grim", "-g", geometry, "-")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("grim failed: %w, stderr: %s", err, stderr.String())
	}

	// Decode PNG
	img, err := png.Decode(&stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	return img, nil
}

// TakeScreenshot captures a screenshot and saves it to a file
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, string, error) {
	// Ensure screenshot directory exists
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	// Find window to get full title
	node, err := findNodeByTitle(windowTitle)
	if err != nil {
		return "", "", err
	}
	fullTitle := node.Name

	img, err := CaptureWindow(windowTitle)
	if err != nil {
		return "", fullTitle, err
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.png", screenshotPrefix, timestamp)
	fullPath := filepath.Join(screenshotDir, filename)

	// Save to file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fullTitle, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fullTitle, fmt.Errorf("failed to encode image: %w", err)
	}

	return fullPath, fullTitle, nil
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
