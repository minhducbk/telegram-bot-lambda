GO=$(shell which go)

build:
	GOARCH=amd64 GOOS=linux ${GO} build -ldflags="-d -s -w" -o main main.go

zip_main:
	GOARCH=amd64 GOOS=linux ${GO} build main.go
	chmod +x main
	zip main.zip main
