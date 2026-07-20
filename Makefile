VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X github.com/tavp-stack/tavpbox/cmd.Version=$(VERSION)"
BINARY := tavpbox

.PHONY: build clean cross test lint cert

build:
	go build $(LDFLAGS) -o $(BINARY).exe .

clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist/

test:
	go test ./...

lint:
	go vet ./...

# Generate cert and embed in binary
cert:
	@echo "Generating wildcard cert for *.tavp.my.id..."
	@mkdir -p internal/certs/embedded
	@if [ -f "$$HOME/.tavpbox/certs/tavp.my.id.pem" ]; then \
		cp $$HOME/.tavpbox/certs/tavp.my.id.pem internal/certs/embedded/; \
		cp $$HOME/.tavpbox/certs/tavp.my.id-key.pem internal/certs/embedded/; \
		echo "Cert copied from ~/.tavpbox/certs/"; \
	else \
		echo "No cert found. Run: tavpbox setup first"; \
		exit 1; \
	fi

# Build and push base images
IMAGE_PREFIX ?= ghcr.io/tavp-stack/tavpbox

release: cross
	@echo "Release binaries in dist/"

cross: clean
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .

# Build and push base images
IMAGE_PREFIX ?= ghcr.io/tavp-stack/tavpbox

images-php:
	podman build -t $(IMAGE_PREFIX)-php:latest -f images/php/Containerfile images/php/

images-node:
	podman build -t $(IMAGE_PREFIX)-node:latest -f images/node/Containerfile images/node/

images-go:
	podman build -t $(IMAGE_PREFIX)-go:latest -f images/go/Containerfile images/go/

images-python:
	podman build -t $(IMAGE_PREFIX)-python:latest -f images/python/Containerfile images/python/

images-all: images-php images-node images-go images-python

images-push: images-all
	podman push $(IMAGE_PREFIX)-php:latest
	podman push $(IMAGE_PREFIX)-node:latest
	podman push $(IMAGE_PREFIX)-go:latest
	podman push $(IMAGE_PREFIX)-python:latest

install: build
	cp $(BINARY).exe $(LOCALAPPDATA)/tavpbox/$(BINARY).exe 2>/dev/null || \
	cp $(BINARY) /usr/local/bin/$(BINARY) 2>/dev/null || \
	echo "Copy manually to PATH"

zip: cross
	cd dist && for f in *; do zip "$$f.zip" "$$f"; done
