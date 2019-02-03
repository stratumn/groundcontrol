all: deps build

deps:
	cd ui && yarn install
	go mod download

gen-ui:
	cd ui && yarn gen

gen-go:
	go generate -tags=release ./...
	gofmt -w $(shell find . -name "auto_*.go")

build-ui:
	cd ui && yarn build

build-go:
	go build -tags "release"

build: gen-ui build-ui gen-go build-go

install:
	cp groundcontrol $$GOPATH/bin

clean-generated:
	rm -f $(shell find . -name "auto_*.go")
	rm -rf $(shell find ui/src -name "__generated__")

clean: clean-generated
	rm -f groundcontrol
	rm -rf dist
	rm -rf ui/node_modules
	rm -rf ui/build

test:
	go test ./...

.PHONY: deps gen-ui gen-go test build-ui build install clean-generated clean
