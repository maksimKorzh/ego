# Build Linux & Windows binaries under linux
export GOOS=linux
go build -o ego e.go
export GOOS=windows
go build -o ego.exe e.go