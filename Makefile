all: deps generate build

deps:
	dep ensure
	cd ui && yarn install

generate:
	cd ui && yarn gen
	go generate -tags "release" ./...

build:
	cd ui && yarn build
	go build -tags "release"

install:
	cp groundcontrol $$GOPATH/bin

.PHONY: deps generate build
