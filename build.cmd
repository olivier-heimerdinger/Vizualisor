@echo off
echo === Vizualisor Build (Windows) ===
set PATH=C:\msys64\ucrt64\bin;%PATH%
set CGO_ENABLED=1
echo Building...
go build -o vizualisor.exe .
if %ERRORLEVEL% EQU 0 (
    echo Build OK : vizualisor.exe
) else (
    echo Build FAILED
    exit /b 1
)
