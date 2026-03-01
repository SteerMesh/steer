# SteerMesh Steer CLI — Makefile
.PHONY: build test lint

BINARY := steer
MAIN   := ./cmd/steer

build:
	go build -o $(BINARY) $(MAIN)

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	go vet ./...
	command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || true
