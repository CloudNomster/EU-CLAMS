//go:build windows
// +build windows

package screenshot

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	gdi32                      = syscall.NewLazyDLL("gdi32.dll")
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
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
	procSetLastError           = kernel32.NewProc("SetLastError")
	procGetLastError           = kernel32.NewProc("GetLastError")
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
		clearLastError()
		ret, _, _ := procGetWindowTextW.Call(
			uintptr(h),
			uintptr(unsafe.Pointer(&title[0])),
			uintptr(len(title)),
		)

		// Convert to Go string and compare
		titleStr := ""
		if ret > 0 {
			titleStr = syscall.UTF16ToString(title[:ret])
		}

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
	clearLastError()
	ret, _, _ := procEnumWindows.Call(cb, 0)
	if ret == 0 {
		// EnumWindows failed
		return 0, "", fmt.Errorf("EnumWindows failed: %s", getLastErrorDetails())
	}

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
	// Log diagnostic information about the window
	fmt.Printf("[DEBUG CaptureWindow] Starting screenshot capture for window handle: %v\n", hwnd)
	fmt.Printf("[DEBUG CaptureWindow] OS Version: %s\n", getWindowsVersionInfo())

	// Get window class name for diagnostics
	var className [256]uint16
	procGetClassNameW := user32.NewProc("GetClassNameW")
	clearLastError()
	classLen, _, _ := procGetClassNameW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&className[0])),
		uintptr(len(className)),
	)
	windowClass := ""
	if classLen > 0 {
		windowClass = syscall.UTF16ToString(className[:classLen])
	} else {
		fmt.Printf("[DEBUG] GetClassNameW failed: %s\n", getLastErrorDetails())
	}
	fmt.Printf("[DEBUG CaptureWindow] Window class: %s\n", windowClass)

	// Get window rectangle
	var rect RECT
	clearLastError()
	ret, _, _ := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get window rectangle: %s", getLastErrorDetails())
	}

	// Calculate width and height
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)

	fmt.Printf("[DEBUG CaptureWindow] Window dimensions: %dx%d (Rect: Left=%d, Top=%d, Right=%d, Bottom=%d)\n",
		width, height, rect.Left, rect.Top, rect.Right, rect.Bottom)

	// Sanity check dimensions
	if width <= 0 || height <= 0 || width > 8192 || height > 8192 {
		return nil, fmt.Errorf("invalid window dimensions: %dx%d", width, height)
	}

	// Get window DC
	clearLastError()
	hdc, _, _ := procGetDC.Call(uintptr(hwnd))
	if hdc == 0 {
		return nil, fmt.Errorf("failed to get window DC: %s", getLastErrorDetails())
	}
	defer procReleaseDC.Call(uintptr(hwnd), hdc)

	// Create compatible DC for bitmap
	clearLastError()
	hdcMem, _, _ := procCreateCompatibleDC.Call(hdc)
	if hdcMem == 0 {
		return nil, fmt.Errorf("failed to create compatible DC: %s", getLastErrorDetails())
	}
	defer procDeleteDC.Call(hdcMem)

	// Create compatible bitmap
	clearLastError()
	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if hBitmap == 0 {
		return nil, fmt.Errorf("failed to create compatible bitmap: %s", getLastErrorDetails())
	}
	defer procDeleteObject.Call(hBitmap)

	// Select bitmap into DC
	clearLastError()
	prevObj, _, _ := procSelectObject.Call(hdcMem, hBitmap)
	if prevObj == 0 {
		fmt.Printf("[DEBUG] SelectObject failed: %s\n", getLastErrorDetails())
	}
	defer procSelectObject.Call(hdcMem, prevObj)
	// Print OS version info to check differences between Windows 10 and 11
	osVersion := getWindowsVersionInfo()
	fmt.Printf("[DEBUG PrintWindow] OS Version: %s\n", osVersion)

	// Try PrintWindow first with full content rendering
	fmt.Printf("[DEBUG PrintWindow] Trying with PW_RENDERFULLCONTENT flag\n")
	clearLastError()
	ret, _, _ = procPrintWindow.Call(
		uintptr(hwnd),
		hdcMem,
		PW_RENDERFULLCONTENT)

	errDetails := getLastErrorDetails()
	fmt.Printf("[DEBUG PrintWindow] PW_RENDERFULLCONTENT result: %d, Error: %s\n", ret, errDetails)

	// If it fails, try regular PrintWindow
	if ret == 0 {
		fmt.Printf("[DEBUG PrintWindow] Trying with no flags\n")
		clearLastError()
		ret, _, _ = procPrintWindow.Call(
			uintptr(hwnd),
			hdcMem,
			0)

		errDetails = getLastErrorDetails()
		fmt.Printf("[DEBUG PrintWindow] No flags result: %d, Error: %s\n", ret, errDetails)

		// If that also fails, try BitBlt as a last resort
		if ret == 0 {
			fmt.Printf("[DEBUG PrintWindow] Trying BitBlt as last resort\n")
			clearLastError()
			ret, _, _ = procBitBlt.Call(
				hdcMem, 0, 0, uintptr(width), uintptr(height),
				hdc, 0, 0, SRCCOPY)

			errDetails = getLastErrorDetails()
			fmt.Printf("[DEBUG BitBlt] Result: %d, Error: %s\n", ret, errDetails)

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

	// Log detailed debug info before GetDIBits call
	logDebugInfo("Before GetDIBits", hwnd, hdc, hdcMem, hBitmap, width, height, bmi)
	fmt.Printf("[DEBUG] About to call GetDIBits with img.Pix length: %d\n", len(img.Pix))

	// Try using our debug wrapper for GetDIBits
	ret, err2 := debugGetDIBits(hdcMem, hBitmap, height, img.Pix, &bmi)
	if ret == 0 {
		// Try alternative approaches if standard approach fails

		// Approach 1: Try with 24-bit color depth instead of 32-bit		fmt.Printf("[DEBUG] Standard GetDIBits failed, checking bitmap info and trying with 24-bit color depth...\n")
		// Try to get actual bitmap info first
		procGetObject := gdi32.NewProc("GetObjectW")
		var bitmapInfo struct {
			Type       int32
			Width      int32
			Height     int32
			WidthBytes int32
			Planes     uint16
			BitsPixel  uint16
			Bits       uintptr
		}
		clearLastError()
		objRet, _, _ := procGetObject.Call(
			hBitmap,
			uintptr(unsafe.Sizeof(bitmapInfo)),
			uintptr(unsafe.Pointer(&bitmapInfo)))

		if objRet != 0 {
			fmt.Printf("[DEBUG] GetObject reports: Width=%d, Height=%d, BitsPixel=%d\n",
				bitmapInfo.Width, bitmapInfo.Height, bitmapInfo.BitsPixel)
		} else {
			fmt.Printf("[DEBUG] GetObject failed: %s\n", getLastErrorDetails())
		}

		// Try with 24-bit format
		bmi.BmiHeader.BiBitCount = 24
		bmi.BmiHeader.BiSizeImage = uint32(width * height * 3) // 3 bytes per pixel

		// Create a new buffer for 24-bit data
		buffer24 := make([]byte, width*height*3)
		clearLastError()
		ret, _, _ = procGetDIBits.Call(
			hdcMem, hBitmap,
			0, uintptr(height),
			uintptr(unsafe.Pointer(&buffer24[0])),
			uintptr(unsafe.Pointer(&bmi)),
			DIB_RGB_COLORS)

		errDetails := getLastErrorDetails()
		fmt.Printf("[DEBUG] 24-bit GetDIBits result: %d, Error: %s\n", ret, errDetails)

		// Approach 2: Try GetBitmapBits as an alternative
		if ret == 0 {
			fmt.Printf("[DEBUG] Trying GetBitmapBits as alternative...\n")
			procGetBitmapBits := gdi32.NewProc("GetBitmapBits")
			bitsSize := int32(width * height * 4)
			bits := make([]byte, bitsSize)
			clearLastError()
			ret, _, _ := procGetBitmapBits.Call(hBitmap, uintptr(bitsSize), uintptr(unsafe.Pointer(&bits[0])))

			errDetails := getLastErrorDetails()
			fmt.Printf("[DEBUG] GetBitmapBits result: %d, Error: %s\n", ret, errDetails)

			if ret > 0 {
				// Copy the bitmap bits to our image
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						i := (y*width + x) * 4
						img.Pix[i] = bits[i+2]   // R
						img.Pix[i+1] = bits[i+1] // G
						img.Pix[i+2] = bits[i]   // B
						img.Pix[i+3] = 255       // A
					}
				}
				fmt.Printf("[DEBUG] Successfully copied bitmap data using GetBitmapBits\n")
			} else {
				return nil, fmt.Errorf("failed to get bitmap bits: %v (width=%d, height=%d, required buffer=%d bytes, image size=%d)", err2, width, height, requiredBufferSize, len(img.Pix))
			}
		}

		if ret == 0 {
			return nil, fmt.Errorf("all bitmap capture methods failed: %v (width=%d, height=%d, required buffer=%d bytes, image size=%d)", err2, width, height, requiredBufferSize, len(img.Pix))
		}
	}

	// Log post-GetDIBits information
	fmt.Printf("[DEBUG] After GetDIBits, ret=%d\n", ret)

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

// getWindowsVersionInfo gets basic Windows version information
func getWindowsVersionInfo() string {
	// Define the structure for Windows version info
	type osVersionInfoEx struct {
		dwOSVersionInfoSize uint32
		dwMajorVersion      uint32
		dwMinorVersion      uint32
		dwBuildNumber       uint32
		dwPlatformId        uint32
		szCSDVersion        [128]uint16
		wServicePackMajor   uint16
		wServicePackMinor   uint16
		wSuiteMask          uint16
		wProductType        byte
		wReserved           byte
	}

	// Load ntdll.dll for RtlGetVersion
	ntdll := syscall.NewLazyDLL("ntdll.dll")
	rtlGetVersion := ntdll.NewProc("RtlGetVersion")

	var osInfo osVersionInfoEx
	osInfo.dwOSVersionInfoSize = uint32(unsafe.Sizeof(osInfo))

	ret, _, _ := rtlGetVersion.Call(uintptr(unsafe.Pointer(&osInfo)))

	if ret != 0 {
		// If RtlGetVersion fails, return a basic message
		return "Failed to get Windows version info"
	}

	return fmt.Sprintf("Windows %d.%d (Build %d)",
		osInfo.dwMajorVersion, osInfo.dwMinorVersion, osInfo.dwBuildNumber)
}

// clearLastError clears the last error before making a Windows API call
func clearLastError() {
	procSetLastError.Call(0)
}

// getLastErrorCode gets the last error code directly
func getLastErrorCode() uint32 {
	ret, _, _ := procGetLastError.Call()
	return uint32(ret)
}

// getLastErrorDetails gets detailed error information from GetLastError
func getLastErrorDetails() string {
	code := getLastErrorCode()

	if code == 0 {
		return "No error (code 0)"
	}

	// Format message buffer
	var buffer [512]uint16
	const FORMAT_MESSAGE_FROM_SYSTEM = 0x00001000
	formatMessage := kernel32.NewProc("FormatMessageW")

	ret, _, _ := formatMessage.Call(
		FORMAT_MESSAGE_FROM_SYSTEM,
		0,
		uintptr(code),
		0,
		uintptr(unsafe.Pointer(&buffer[0])),
		512,
		0)

	if ret == 0 {
		return fmt.Sprintf("Error code %d: (Unable to format error message)", code)
	}

	// Convert to Go string and trim
	msg := syscall.UTF16ToString(buffer[:])
	return fmt.Sprintf("Error code %d: %s", code, strings.TrimSpace(msg))
}

// logDebugInfo logs detailed debug information about the window and bitmap
func logDebugInfo(prefix string, hwnd syscall.Handle, hdc, hdcMem, hBitmap uintptr, width, height int, bmi BITMAPINFO) {
	// Check if this is Windows 10 or 11
	osVersion := getWindowsVersionInfo()
	fmt.Printf("[DEBUG %s] OS Version: %s\n", prefix, osVersion)

	// Get window class
	var className [256]uint16
	procGetClassNameW := user32.NewProc("GetClassNameW")
	clearLastError()
	ret, _, _ := procGetClassNameW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&className[0])),
		uintptr(len(className)),
	)
	class := ""
	if ret > 0 {
		class = syscall.UTF16ToString(className[:ret])
	} else {
		fmt.Printf("[DEBUG %s] GetClassNameW failed: %s\n", prefix, getLastErrorDetails())
	}
	fmt.Printf("[DEBUG %s] Window class: %s\n", prefix, class)

	// Check DPI information
	procGetDpiForWindow := user32.NewProc("GetDpiForWindow")
	if procGetDpiForWindow.Find() == nil {
		clearLastError()
		dpi, _, _ := procGetDpiForWindow.Call(uintptr(hwnd))
		if dpi > 0 {
			fmt.Printf("[DEBUG %s] Window DPI: %d\n", prefix, dpi)
		} else {
			fmt.Printf("[DEBUG %s] GetDpiForWindow failed: %s\n", prefix, getLastErrorDetails())
		}
	} else {
		fmt.Printf("[DEBUG %s] GetDpiForWindow not available\n", prefix)
	}

	// Log bitmap info
	fmt.Printf("[DEBUG %s] Bitmap info: %dx%d, bitCount=%d, compression=%d, sizeImage=%d\n",
		prefix, width, height, bmi.BmiHeader.BiBitCount, bmi.BmiHeader.BiCompression, bmi.BmiHeader.BiSizeImage)

	// Log memory info for image buffer
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	fmt.Printf("[DEBUG %s] Memory stats: Alloc=%d MB, Sys=%d MB\n",
		prefix, memStats.Alloc/1024/1024, memStats.Sys/1024/1024)
}

// debugGetDIBits is a wrapper around GetDIBits with extensive diagnostics
func debugGetDIBits(hdcMem, hBitmap uintptr, height int, imgPix []byte, bmi *BITMAPINFO) (uintptr, error) {
	// Print debug info before GetDIBits
	fmt.Printf("[DEBUG GetDIBits] BEFORE call: imgPix len=%d, bmiHeader size=%d\n",
		len(imgPix), bmi.BmiHeader.BiSize)

	// Try with different start scan line values
	fmt.Printf("[DEBUG GetDIBits] Trying with standard parameters...\n")
	clearLastError()
	ret, _, err := procGetDIBits.Call(
		hdcMem, hBitmap,
		0, uintptr(height),
		uintptr(unsafe.Pointer(&imgPix[0])),
		uintptr(unsafe.Pointer(bmi)),
		DIB_RGB_COLORS)

	// Check error and print details
	errDetails := getLastErrorDetails()
	fmt.Printf("[DEBUG GetDIBits] Return value: %d, Error: %s\n", ret, errDetails)

	// Try alternate strategies if standard call fails
	if ret == 0 {
		// Try with different scan parameters
		fmt.Printf("[DEBUG GetDIBits] Trying alternate parameters...\n")

		// Try with a small sleep to allow for any async operations
		time.Sleep(50 * time.Millisecond)

		clearLastError()
		ret, _, err = procGetDIBits.Call(
			hdcMem, hBitmap,
			0, uintptr(height),
			uintptr(unsafe.Pointer(&imgPix[0])),
			uintptr(unsafe.Pointer(bmi)),
			DIB_RGB_COLORS)

		errDetails = getLastErrorDetails()
		fmt.Printf("[DEBUG GetDIBits] Alternate params result: %d, Error: %s\n", ret, errDetails)
	}

	return ret, err
}
