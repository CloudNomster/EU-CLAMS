@echo off
echo Testing partial window title matching...
REM First argument is the window title prefix, default is "Entropia Universe Client"
if "%1"=="" (
    set WINDOW_TITLE=Entropia Universe Client
) else (
    set WINDOW_TITLE=%1
)

echo Searching for windows with title starting with: %WINDOW_TITLE%
echo.
go run cmd\test_screenshot\test_capture.go "%WINDOW_TITLE%"
if %ERRORLEVEL% NEQ 0 (
    echo Test failed!
    exit /b 1
)
echo Test completed successfully!
echo Check the data\screenshots\test directory for the captured screenshot.
