VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X github.com/tavp-stack/tavpbox/cmd.Version=$(VERSION)"
BINARY := tavpbox

.PHONY: build clean cross test lint

build:
	go build $(LDFLAGS) -o $(BINARY).exe .

clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist/

test:
	go test ./...

lint:
	go vet ./...

cross: clean
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .

install: build
	cp $(BINARY).exe $(LOCALAPPDATA)/tavpbox/$(BINARY).exe 2>/dev/null || \
	cp $(BINARY) /usr/local/bin/$(BINARY) 2>/dev/null || \
	echo "Copy manually to PATH"

release: cross
	@echo "Release binaries in dist/"
	@ls -la dist/

zip: cross
	cd dist && for f in *; do zip "$$f.zip" "$$f"; done
