all:
	cd ui && yarn build
	go generate -tags "release" ./...
	go build -tags "release" -o ./build/groundcontrol ./server
