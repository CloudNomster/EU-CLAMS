//go:build windows
// +build windows

package screenshot

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	gdi32                      = syscall.NewLazyDLL("gdi32.dll")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procGetDC                  = user32.NewProc("GetDC")
	procReleaseDC              = user32.NewProc("ReleaseDC")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
	procPrintWindow            = user32.NewProc("PrintWindow")
	procIsWindowVisible        = user32.NewProc("IsWindowVisible")
)

// RECT is the Windows RECT structure
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// BITMAPINFOHEADER is the Windows BITMAPINFOHEADER structure
type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// BITMAPINFO is the Windows BITMAPINFO structure
type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD
}

// RGBQUAD is the Windows RGBQUAD structure
type RGBQUAD struct {
	Blue     byte
	Green    byte
	Red      byte
	Reserved byte
}

const (
	// RGB format constants
	DIB_RGB_COLORS = 0
	SRCCOPY        = 0x00CC0020
	BI_RGB         = 0
)

const (
	// PrintWindow options
	PW_CLIENTONLY        = 0x00000001
	PW_RENDERFULLCONTENT = 0x00000002
)

// FindWindowWithPartialTitle finds a window where the title starts with the given prefix
func FindWindowWithPartialTitle(titlePrefix string) (syscall.Handle, error) {
	handle, _, err := findWindowWithPartialTitleAndGetTitle(titlePrefix)
	return handle, err
}

// findWindowWithPartialTitleAndGetTitle finds a window where the title starts with the given prefix and returns both handle and full title
func findWindowWithPartialTitleAndGetTitle(titlePrefix string) (syscall.Handle, string, error) {
	// Define the callback function for EnumWindows
	var hwnd syscall.Handle
	var fullWindowTitle string

	// We need to convert the Go string to a format we can compare with Windows titles
	titlePrefixLower := strings.ToLower(titlePrefix)

	// Use EnumWindows to find the window with the title starting with our prefix
	cb := syscall.NewCallback(func(h syscall.Handle, param uintptr) uintptr {
		// Get the window title
		var title [256]uint16
		procGetWindowTextW := user32.NewProc("GetWindowTextW")
		procGetWindowTextW.Call(
			uintptr(h),
			uintptr(unsafe.Pointer(&title[0])),
			uintptr(len(title)),
		)

		// Convert to Go string and compare
		titleStr := syscall.UTF16ToString(title[:])
		if strings.HasPrefix(strings.ToLower(titleStr), titlePrefixLower) && titleStr != "" {
			// Found a matching window
			hwnd = h
			fullWindowTitle = titleStr
			return 0 // Stop enumeration
		}

		return 1 // Continue enumeration
	})

	// Enumerate all windows
	procEnumWindows := user32.NewProc("EnumWindows")
	procEnumWindows.Call(cb, 0)

	if hwnd == 0 {
		return 0, "", fmt.Errorf("no window with title prefix '%s' found", titlePrefix)
	}

	return hwnd, fullWindowTitle, nil
}

// CaptureWindow takes a screenshot of the specified window by title
func CaptureWindow(windowTitle string) (image.Image, error) {
	// Find the window handle using partial title match
	hwnd, err := FindWindowWithPartialTitle(windowTitle)
	if err != nil {
		return nil, err
	}

	// Get window rectangle
	var rect RECT
	ret, _, err := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get window rectangle: %v", err)
	}

	// Calculate width and height
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)

	// Sanity check dimensions
	if width <= 0 || height <= 0 || width > 8192 || height > 8192 {
		return nil, fmt.Errorf("invalid window dimensions: %dx%d", width, height)
	}

	// Get window DC
	hdc, _, _ := procGetDC.Call(uintptr(hwnd))
	if hdc == 0 {
		return nil, fmt.Errorf("failed to get window DC")
	}
	defer procReleaseDC.Call(uintptr(hwnd), hdc)

	// Create compatible DC for bitmap
	hdcMem, _, _ := procCreateCompatibleDC.Call(hdc)
	if hdcMem == 0 {
		return nil, fmt.Errorf("failed to create compatible DC")
	}
	defer procDeleteDC.Call(hdcMem)

	// Create compatible bitmap
	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if hBitmap == 0 {
		return nil, fmt.Errorf("failed to create compatible bitmap")
	}
	defer procDeleteObject.Call(hBitmap)

	// Select bitmap into DC
	prevObj, _, _ := procSelectObject.Call(hdcMem, hBitmap)
	defer procSelectObject.Call(hdcMem, prevObj)

	// Try PrintWindow first with full content rendering
	ret, _, _ = procPrintWindow.Call(
		uintptr(hwnd),
		hdcMem,
		PW_RENDERFULLCONTENT)

	// If it fails, try regular PrintWindow
	if ret == 0 {
		ret, _, _ = procPrintWindow.Call(
			uintptr(hwnd),
			hdcMem,
			0)

		// If that also fails, try BitBlt as a last resort
		if ret == 0 {
			ret, _, _ = procBitBlt.Call(
				hdcMem, 0, 0, uintptr(width), uintptr(height),
				hdc, 0, 0, SRCCOPY)
			if ret == 0 {
				return nil, fmt.Errorf("all screen capture methods failed")
			}
		}
	}
	// Create Go image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Calculate the required buffer size explicitly
	// Each pixel needs 4 bytes (RGBA), so total size is width * height * 4
	requiredBufferSize := width * height * 4

	// Ensure our image buffer is large enough
	if len(img.Pix) < requiredBufferSize {
		return nil, fmt.Errorf("image buffer too small: got %d bytes, need %d bytes", len(img.Pix), requiredBufferSize)
	}

	// Prepare BITMAPINFO structure
	bmi := BITMAPINFO{}
	bmi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmi.BmiHeader))
	bmi.BmiHeader.BiWidth = int32(width)
	bmi.BmiHeader.BiHeight = -int32(height) // Negative for top-down
	bmi.BmiHeader.BiPlanes = 1
	bmi.BmiHeader.BiBitCount = 32
	bmi.BmiHeader.BiCompression = BI_RGB
	bmi.BmiHeader.BiSizeImage = uint32(requiredBufferSize) // Use the explicitly calculated size

	// Try alternative methods if the first GetDIBits call fails
	ret, _, _ = procGetDIBits.Call(
		hdcMem, hBitmap,
		0, uintptr(height),
		uintptr(unsafe.Pointer(&img.Pix[0])),
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS)
	if ret == 0 {
		// Try a different approach with separate buffer allocation
		// Use the previously calculated buffer size to ensure consistency
		buffer := make([]byte, requiredBufferSize)

		ret, _, lastErr := procGetDIBits.Call(
			hdcMem, hBitmap,
			0, uintptr(height),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(unsafe.Pointer(&bmi)),
			DIB_RGB_COLORS)

		if ret == 0 {
			return nil, fmt.Errorf("failed to get DIB bits: %v (width=%d, height=%d, required buffer=%d bytes, buffer size=%d)", lastErr, width, height, requiredBufferSize, len(buffer))
		}

		// Copy from buffer to image
		if len(buffer) > len(img.Pix) {
			return nil, fmt.Errorf("buffer size mismatch: buffer=%d bytes, image=%d bytes", len(buffer), len(img.Pix))
		}
		copy(img.Pix, buffer)
	}

	// Fix color channel order: Windows GDI returns BGR but Go expects RGB
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+2] = img.Pix[i+2], img.Pix[i] // Swap R and B channels
	}

	return img, nil
}

// TakeScreenshot captures a screenshot of the Entropia Universe client window and saves it to the specified directory
// It returns the saved screenshot path and the full window title
func TakeScreenshot(windowTitle, screenshotDir, screenshotPrefix string) (string, string, error) {
	// Ensure screenshot directory exists
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	// First get the full window title and handle
	hwnd, fullTitle, err := findWindowWithPartialTitleAndGetTitle(windowTitle)
	if err != nil {
		return "", "", fmt.Errorf("failed to find window: %w", err)
	}

	// Check if the window is visible before taking a screenshot
	if !IsWindowVisible(hwnd) {
		return "", fullTitle, fmt.Errorf("window is not visible: %s", fullTitle)
	}

	// Capture window
	img, err := CaptureWindow(windowTitle)
	if err != nil {
		return "", fullTitle, fmt.Errorf("failed to capture window: %w", err)
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

	// Encode as PNG
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

// GetFullWindowTitle finds the window with the given title prefix and returns its full title
func GetFullWindowTitle(windowTitle string) (string, error) {
	_, fullTitle, err := findWindowWithPartialTitleAndGetTitle(windowTitle)
	if err != nil {
		return "", err
	}
	return fullTitle, nil
}

// IsWindowVisible checks if the window with the given handle is visible
func IsWindowVisible(hwnd syscall.Handle) bool {
	ret, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return ret != 0
}
