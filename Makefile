GO ?= go
.DEFAULT_GOAL := build
.PHONY: build test vet lint langsmoke race cover generate provenance check
build:
	$(GO) build ./...
test:
	$(GO) test -timeout 20m ./...
vet:
	$(GO) vet ./...
lint:
	golangci-lint run
	govulncheck ./...
langsmoke:
	cd tools/langsmoke && GOWORK=off $(GO) run .
race:
	$(GO) test -race ./kustolang ./kustolang/engine ./kustolang/environment
cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1
generate:
	$(GO) generate ./kustolang
provenance:
	$(GO) run ./cmd/kusto-upstream-check -verify-provenance -local-only
check: test vet provenance langsmoke
