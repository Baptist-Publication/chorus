.PHONY: chorus chorustool test all
all: chorus chorustool test

chorus:
	go build  -ldflags "-X github.com/Baptist-Publication/chorus/src/chain/version.commitVer=`git rev-parse HEAD`" -o ./build/chorus ./src/chain
chorustool:
	go build -ldflags "-X github.com/Baptist-Publication/chorus/src/client/main.version=`git rev-parse HEAD`" -o ./build/chorustool ./src/client
test:
	go test ./src/tools/state
	go test ./src/tools
	go test ./src/chain/node/
