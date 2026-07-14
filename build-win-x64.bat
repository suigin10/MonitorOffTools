@echo off
setlocal
cd /d "%~dp0"

echo Building MonitorOffTools v1.0.0...
where go >nul 2>nul
if errorlevel 1 (
  echo Go is not installed or not available in PATH.
  pause
  exit /b 1
)

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -buildvcs=false -trimpath -ldflags="-H windowsgui -s -w" -o MonitorOffTools.exe .

if errorlevel 1 (
  echo.
  echo Build failed.
  pause
  exit /b 1
)

echo.
echo Build complete: MonitorOffTools.exe
pause
