GO_AUTO = $(shell find . -name 'auto_*.go')
UI_AUTO = $(shell find ../groundcontrol-ui/src -name '__generated__')
GARBAGE := groundcontrol dist electron/assets/vendor ../groundcontrol-ui/node_modules ../groundcontrol-ui/build 

all: deps build

deps-ui:
	cd ../groundcontrol-ui && yarn install

deps-go:
	go mod download

deps: deps-ui deps-go

gen-ui:
	cd ../groundcontrol-ui && yarn gen

gen-go-release:
	go generate -tags=release ./...

gen-go-fmt:
	gofmt -w $(GO_AUTO)

gen-go: gen-go-release gen-go-fmt

build-ui:
	cd ../groundcontrol-ui && yarn build

build-go:
	go build -tags "release"

build: gen-ui build-ui gen-go build-go

install:
	cp groundcontrol $$GOPATH/bin

clean-generated:
	rm -rf $(GO_AUTO) $(UI_AUTO)

clean: clean-generated
	rm -rf $(GARBAGE)

test-ui:
	cd ../groundcontrol-ui && CI=1 yarn test

test-go:
	go test ./...

test: test-ui test-go

.PHONY: deps deps-ui deps-go gen-ui gen-go gen-go-release gen-go-fmt test-ui test-go test build-ui build install clean-generated clean
