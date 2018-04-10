.PHONY: chorus chorustool all
all: chorus chorustool

chorus:
	go build  -ldflags "-X github.com/Baptist-Publication/chorus/src/chain/version.commitVer=`git rev-parse HEAD`" -o ./build/chorus ./src/chain
chorustool:
	go build -ldflags "-X github.com/Baptist-Publication/chorus/src/client/main.version=`git rev-parse HEAD`" -o ./build/chorustool ./src/client
test:
	go test ./src/tools/state
proto:
	protoc --proto_path=$(GOPATH)/src --proto_path=src/chain/app/remote --go_out=plugins=grpc:src/chain/app/remote src/chain/app/remote/*.proto
	protoc --proto_path=$(GOPATH)/src --proto_path=src/example/types --go_out=plugins=grpc:src/example/types src/example/types/*.proto
	#protoc --proto_path=$(GOPATH)/src --proto_path=src/chain/node/protos --gofast_out=plugins=grpc:src/chain/node/protos src/chain/node/protos/*.proto
	protoc --proto_path=src/types --go_out=src/types src/types/*.proto
