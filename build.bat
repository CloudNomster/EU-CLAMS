@echo off
REM Build script for EU-CLAMS

REM Add GCC to PATH
set PATH=C:\TDM-GCC-64\bin;%PATH%

REM Check if GCC is installed
where gcc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: GCC not found in PATH even after adding C:\TDM-GCC-64\bin
    echo Please make sure TDM-GCC is properly installed.
    exit /b 1
)

REM Set CGO_ENABLED for GUI support
set CGO_ENABLED=1

REM Build the application
echo Building EU-CLAMS...
go build -o eu-clams.exe ./cmd/app

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b 1
)

echo Build successful!
echo To run the application with GUI: eu-tool.exe
