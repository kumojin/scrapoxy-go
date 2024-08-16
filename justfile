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

buildFingerprintServer:
  go build -o bin/fingerprint-server ./fingerprint-server

buildFingerprintServerLinux $CGO_ENABLED="0":
  GOOS=linux GOARCH=amd64 go build -o bin/fingerprint-server-linux-amd64 ./fingerprint-server
  GOOS=linux GOARCH=arm64 go build -o bin/fingerprint-server-linux-arm64 ./fingerprint-server


buildAll: buildProxy buildProxyDispatcher buildFingerprintServer

buildAllLinux: buildProxyDispatcherLinux buildProxyLinux buildFingerprintServerLinux
