@echo off
REM TAVPBox - Global wrapper script
REM This file should be placed in a directory that's in your PATH

setlocal

REM Get the directory where this script is located
set "SCRIPT_DIR=%~dp0"

REM Run the actual binary
"%SCRIPT_DIR%tavpbox.exe" %*

endlocal
