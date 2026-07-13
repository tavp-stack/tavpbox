BINARY=tavpbox
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")

.PHONY: build install clean cross test

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) .

install: build
	sudo install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

cross:
	@echo "Building for all platforms..."
	GOOS=linux  GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux  GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-windows-amd64.exe .
	@echo "✓ Binaries in dist/"

clean:
	rm -f $(BINARY)
	rm -rf dist/

test:
	go test ./... -v -count=1
