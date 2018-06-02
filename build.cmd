@echo off

rem A build script that automatically picks the right library from the subfolders in "libs".
rem Use this script if you are unable or don't want to use the system library.

setlocal

rem Package-specific libraries
set ldargs=-lsquish -lgomp -lm -lstdc++

:ArgsLoop
if "%~1"=="" goto ArgsFinished
echo "%~1" | find /i "--libdir" >nul && goto ArgsLibDir
echo "%~1" | find /i "--help" >nul && goto ArgsHelp
goto ArgsUpdate

:ArgsLibDir
shift
set libdir=%~1
goto ArgsUpdate

:ArgsHelp
echo Usage: %~n0%~x0 [options]
echo.
echo Options:
echo   --libdir path    Override library path
echo   --help           This help
goto Finished

:ArgsUpdate
shift
goto ArgsLoop

:ArgsFinished
if not "%libdir%"=="" goto SkipAuto

rem Autodetect
go env GOARCH | findstr /i "amd64" >nul && goto x86_64

echo Detected: os=windows, arch=386
set libdir=libs/windows/386
goto Continued

:x86_64
echo Detected: os=windows, arch=amd64
set libdir=libs/windows/amd64
goto Continued

:SkipAuto
echo Using libdir: %libdir%

:Continued
echo Building library...
set CGO_LDFLAGS=-L%libdir% %ldargs%
go build && go install && goto Success || goto Failed

:Success
echo Finished.
goto Finished

:Failed
echo Cancelled.
endlocal
exit /b 1

:Finished
endlocal
