SHELL := /bin/bash

BINARY ?= gpcm
PKG := main.go
CONFIG ?= config.yaml

.PHONY: all build tidy fmt test clean run init generate

all: build

build:
	go build -o $(BINARY) $(PKG)

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY) -config $(CONFIG) $(ARGS)

init: build
	./$(BINARY) -config $(CONFIG) init

generate: build
	./$(BINARY) -config $(CONFIG) generate