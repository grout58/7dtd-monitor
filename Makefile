BINARY_NAME=7dtd-monitor

all: build

build:
	go build -o $(BINARY_NAME) cmd/7dtd-monitor/main.go

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-windows-amd64.exe
	rm -f $(BINARY_NAME)-linux-amd64
	rm -f $(BINARY_NAME)-darwin-amd64
	rm -f $(BINARY_NAME)-darwin-arm64

# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 cmd/7dtd-monitor/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe cmd/7dtd-monitor/main.go

build-macos-intel:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 cmd/7dtd-monitor/main.go

build-macos-apple-silicon:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 cmd/7dtd-monitor/main.go

build-all: build-linux build-windows build-macos-intel build-macos-apple-silicon
