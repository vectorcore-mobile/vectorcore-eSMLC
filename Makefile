.PHONY: build test vet

GOCACHE ?= /tmp/vectorcore-esmlc-go-cache

build:
	GOCACHE=$(GOCACHE) CGO_ENABLED=0 go build -buildvcs=false ./...

test:
	GOCACHE=$(GOCACHE) go test -buildvcs=false ./... -count=1

vet:
	GOCACHE=$(GOCACHE) go vet -buildvcs=false ./...
