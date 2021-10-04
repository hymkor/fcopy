@echo off
setlocal
set "PROMPT=$$ "
call :"%1"
endlocal
exit /b

:""
    @echo on
    pushd internal\file & go fmt & popd
    go fmt
    go build
    @echo off
    exit /b

:"pack"
    for /F %%I in ('git.exe describe --tags') do set "VERSION=%%I"
    for %%I in (386 amd64) do call :pack1 "%%~I"
    exit /b

:pack1
    @echo on
    set "GOARCH=%~1"
    call :""
    zip "fcopy-windows-%GOARCH%-%VERSION%.zip" fcopy.exe
    @echo off
    exit /b

:"clean"
    del *.zip
    exit /b

:"install"
    @echo on
    @for %%I in (fcopy.exe) do copy fcopy.exe %%~$PATH:I
    @echo off
    exit /b
