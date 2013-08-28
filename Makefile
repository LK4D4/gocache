GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w

V=4
SIZE=10000

GOPATH := $(shell pwd)

all: install

bench: install
	bin/bench -v $(V) -c $(SIZE)

install:
	GOPATH=$(GOPATH) $(GOINSTALL) ./...

fmt:
	GOPATH=$(GOPATH) $(GOFMT) $(GOPATH)

test:
	GOPATH=$(GOPATH) $(GOTEST) ./...


cleanvim:
	find -name "*.swp" -delete
	find -name "*.swo" -delete

clean: cleanvim
	$(GOCLEAN)
