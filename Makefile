
APP_NAME=gotun
DIR=$(dirname "$0")

GOPATH=$(shell go env GOBIN)

# 生成linux可执行文件
linux:
	GOOS=linux GOARCH=amd64 go build -o ./dist/$(APP_NAME) ./cmd/gotun

openwrt:
	GOOS=linux GOARCH=mipsle go build -o ./dist/$(APP_NAME) ./cmd/gotun
	
# 生成windows可执行文件
windows:
	GOOS=windows GOARCH=amd64 go build -o ./dist/$(APP_NAME).exe ./cmd/gotun

# 默认生成linux可执行文件
default: linux
