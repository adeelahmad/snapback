# Snapback developer Makefile (GNU make). Run `make` or `make help` for targets.

export GOTOOLCHAIN ?= auto

PREFIX  ?= /usr/local
DESTDIR ?=
BINARY  := bin/snapback

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)

VERSION_PKG := github.com/adeelahmad/snapback/internal/version
LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).Target=$(GOOS)/$(GOARCH)

COVERAGE_MIN := 80

.PHONY: help build install uninstall test cover lint fmt fmt-check vet vuln \
	docs release-check ci clean compile lint-actions lint-shell

help: ## List targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/snapback (CGO off) with version info
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/snapback

install: build ## Install bin/snapback to $(DESTDIR)$(PREFIX)/bin
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/snapback

uninstall: ## Remove $(DESTDIR)$(PREFIX)/bin/snapback
	rm -f $(DESTDIR)$(PREFIX)/bin/snapback

test: ## Run all tests with the race detector
	go test -race ./...

cover: ## Run tests with coverage and fail below 80% total
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | awk '/^total:/{gsub("%","",$$NF); if ($$NF+0 < $(COVERAGE_MIN)) { print "coverage " $$NF "% is below the $(COVERAGE_MIN)% minimum"; exit 1 } else { print "coverage " $$NF "% OK (>=$(COVERAGE_MIN)%)" } }'

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format Go code in place
	gofmt -w .
	goimports -w .

fmt-check: ## Fail if Go code is not formatted
	test -z "$$(gofmt -l .)"
	test -z "$$(goimports -l .)"

vet: ## Run go vet
	go vet ./...

vuln: ## Scan dependencies for known vulnerabilities
	govulncheck ./...

docs: ## Build the docs site strictly
	mkdocs build --strict --site-dir site

release-check: ## Validate the GoReleaser config
	goreleaser check

compile: ## Compile every package without CGO
	CGO_ENABLED=0 go build ./...

lint-actions: ## Lint GitHub Actions workflows
	actionlint

lint-shell: ## Lint install.sh as POSIX sh
	shellcheck -s sh install.sh

ci: fmt-check compile vet lint test cover vuln lint-actions lint-shell docs release-check ## Run every standards gate in order

clean: ## Remove build, docs and coverage output
	rm -rf bin/ site/ coverage.out
