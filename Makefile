APP_NAME=gophkeeper
VERSION=0.1.0
BUILD_DATE=$(shell date -u +%Y-%m-%d)

all: build

build:
	go build \
		-ldflags "-X 'github.com/patric-chuzhbe/gophkeeper/internal/version.Version=$(VERSION)' -X 'github.com/patric-chuzhbe/gophkeeper/internal/version.BuildDate=$(BUILD_DATE)'" \
		-o ./cmd/client/client.exe \
		./cmd/client
	go build -ldflags "-X 'github.com/patric-chuzhbe/gophkeeper/internal/version.Version=$(VERSION)' -X 'github.com/patric-chuzhbe/gophkeeper/internal/version.BuildDate=$(BUILD_DATE)'" \
		-o ./cmd/server/server.exe \
		./cmd/server

proto:
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		internal/servergrpc/proto/vault.proto
