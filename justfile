#!/usr/bin/env just --justfile

update:
  go get -u
  go mod tidy -v

buildProxy:
  go build -o bin/proxy ./proxy

buildProxyLinux $CGO_ENABLED="0":
  GOOS=linux GOARCH=amd64 go build -o bin/proxy-linux-amd64 ./proxy
  GOOS=linux GOARCH=arm64 go build -o bin/proxy-linux-arm64 ./proxy

buildProxyDispatcher:
  go build -o bin/proxy-dispatcher ./proxy-dispatcher

buildProxyDispatcherLinux $CGO_ENABLED="0":
  GOOS=linux GOARCH=amd64 go build -o bin/proxy-dispatcher-linux-amd64 ./proxy-dispatcher
  GOOS=linux GOARCH=arm64 go build -o bin/proxy-dispatcher-linux-arm64 ./proxy-dispatcher

buildAll: buildProxy buildProxyDispatcher

buildAllLinux: buildProxyDispatcherLinux buildProxyLinux

