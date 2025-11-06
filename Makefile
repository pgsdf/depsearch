# Simple Makefile for GhostBSD/FreeBSD/PGSD
APP=depsearch
GO?=go
PREFIX?=/usr/local
BINDIR?=$(PREFIX)/bin

all: build

build:
	$(GO) build -trimpath -ldflags "-s -w" -o $(APP) ./cmd/depsearch

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(APP) $(DESTDIR)$(BINDIR)/$(APP)

clean:
	rm -f $(APP)

# Cross builds for common GhostBSD targets
build-amd64:
	GOOS=freebsd GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o $(APP)-freebsd-amd64 ./cmd/depsearch

build-arm64:
	GOOS=freebsd GOARCH=arm64 $(GO) build -trimpath -ldflags "-s -w" -o $(APP)-freebsd-arm64 ./cmd/depsearch

.PHONY: all build install clean build-amd64 build-arm64
