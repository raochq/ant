@echo off
set pwd=%~dp0
set Output=%pwd%bin
set flag=-gcflags "all=-N -l"

if [%1] == [] goto:all

if %1==clean (
call:clean
) else if %1==proto (
call:proto %2
) else if %1==gen (
call:gen
) else if %1==bin (
call:bin
) else (
call:build %1
)
goto:exit

:all
call:proto
call:gen
call:bin
exit /b %errorlevel%

:proto
set PROTO_SRC=protocol/proto
set PROTO_DEST=protocol/pb
protoc -I=%PROTO_SRC% --go_out=%PROTO_DEST% %PROTO_SRC%\*.proto
exit /b

:gen
go generate ./errcode ./protocol
exit /b

:bin
set APPS=app
call:build %APPS%
exit /b

:build
go build -o %Output%\%1.exe %flag% ./%1
exit /b

:clean
go clean
exit /b

:exit