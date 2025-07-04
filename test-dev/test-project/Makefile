# Generated Makefile for test-project

.PHONY: build test clean run install

build:
	go build -v ./...

test:
	go test -v ./...

clean:
	go clean
	rm -f test-project

run:
	go run main.go

install:
	go install

fmt:
	go fmt ./...

vet:
	go vet ./...

mod-tidy:
	go mod tidy

deps:
	go mod download
