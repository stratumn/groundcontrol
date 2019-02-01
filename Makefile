all: deps build

deps:
	cd ui && yarn install

build: clean_generated
	cd ui && yarn gen
	cd ui && yarn build
	go generate -tags "release" ./...
	go build -tags "release"

install:
	cp groundcontrol $$GOPATH/bin

clean_generated:
	rm -f $(shell find . -name "auto_*.go")
	rm -rf $(shell find ui/src -name "__generated__")

clean: clean_generated
	rm -f groundcontrol
	rm -rf ui/node_modules

.PHONY: deps build install clean_generated clean
