.PHONY: build test vet clean

GOCACHE ?= /tmp/vectorcore-esmlc-go-cache
BINARY := esmlc

# `go build ./...` (a package pattern, not a single package) only
# type-checks/compiles for verification and never writes a binary to disk,
# even though cmd/esmlc is a main package — so this also builds the actual
# runnable binary explicitly, which `go build ./...` alone never did.
build:
	GOCACHE=$(GOCACHE) CGO_ENABLED=0 go build -buildvcs=false ./...
	GOCACHE=$(GOCACHE) CGO_ENABLED=0 go build -buildvcs=false -o bin/$(BINARY) ./cmd/esmlc

test:
	GOCACHE=$(GOCACHE) go test -buildvcs=false ./... -count=1

vet:
	GOCACHE=$(GOCACHE) go vet -buildvcs=false ./...

clean:
	rm -f bin/$(BINARY)
