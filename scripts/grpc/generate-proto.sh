#!/usr/bin/env bash

export GO111MODULE=on
go get github.com/golang/protobuf/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.0

## todo : 나의 *.proto 정보
#protoc -I idl/ --go_out=plugins=grpc:protocol idl/*.proto
#protoc -I idl/ idl/presentation_server.proto --go_out=plugins=grpc:./protocol
#protoc -I idl/ idl/session_server.proto --go_out=plugins=grpc:./protocol