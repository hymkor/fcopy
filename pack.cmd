@echo off
setlocal
set "PROMPT=$$ "
set /P "VERSION=VersionNumber ? "
for %%I in (386 amd64) do call :main1 "%%~I"
endlocal
exit /b

:main1
    @echo on
    set "GOARCH=%~1"
    go build
    upx fcopy.exe
    zip "fcopy-windows-%GOARCH%-%VERSION%.zip" fcopy.exe
    @echo off
    exit /b
