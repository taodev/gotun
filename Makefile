
APP_NAME=gotun
DIR=$(dirname "$0")

GOPATH=$(shell go env GOBIN)

# go env
GO111MODULE=on
CGO_ENABLED=0 
GOOS=linux 
GOARCH=amd64

# mkdir -p ./dist

# 生成linux可执行文件
linux:
	rm -fr ./dist
	go build --tags=netgo -o ./dist/$(APP_NAME) ./cmd/gotun

openwrt:
	GOOS=linux GOARCH=mipsle go build -o ./dist/$(APP_NAME) ./cmd/gotun
	
# 生成windows可执行文件
windows:
	GOOS=windows GOARCH=amd64 go build -o ./dist/$(APP_NAME).exe ./cmd/gotun

docker:
	go mod tidy
	go clean -cache

	rm -fr ./dist
	go build --tags=netgo -o ./dist/$(APP_NAME) ./cmd/gotun

	cp -f ./Dockerfile ./dist
	docker build -t taodev/gotun:latest ./dist
	docker push taodev/gotun:latest

clean-docker:
	docker taodev/gotun:latest -f
	docker builder prune -a -f

# 默认生成linux可执行文件
default: linux
