.PHONY: all clean build

BINARY_NAME=blevaettir
GO=go
GOFLAGS=-ldflags="-s -w"

PLATFORMS=linux/amd64 linux/arm64 linux/386

all: build

build: $(PLATFORMS)

linux/amd64:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o bin/$@/$(BINARY_NAME) ./cmd/blevaettir

linux/arm64:
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o bin/$@/$(BINARY_NAME) ./cmd/blevaettir

linux/386:
	GOOS=linux GOARCH=386 $(GO) build $(GOFLAGS) -o bin/$@/$(BINARY_NAME) ./cmd/blevaettir

clean:
	rm -rf bin
