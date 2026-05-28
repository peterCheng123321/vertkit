#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${GOCACHE:-/private/tmp/vertkit-go-cache}"

gofmt_output="$(gofmt -l .)"
if [[ -n "$gofmt_output" ]]; then
  printf '%s\n' "$gofmt_output"
  exit 1
fi

go test ./...
go vet ./...
go test -race ./...
