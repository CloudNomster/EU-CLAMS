@echo off
echo Testing screenshot capture...
go run cmd\test_screenshot\test_capture.go %*
if %ERRORLEVEL% NEQ 0 (
    echo Test failed!
    exit /b 1
)
echo Test completed successfully!
echo Check the data\screenshots\test directory for the captured screenshot.
