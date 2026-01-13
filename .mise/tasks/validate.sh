#!/usr/bin/env bash
set -e

go build ./...
go mod tidy
go vet ./...
golangci-lint run
