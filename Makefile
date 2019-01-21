all:
	cd ui && yarn gen && yarn build
	go generate -tags "release" ./...
	go build -tags "release" -o ./build/groundcontrol ./server
