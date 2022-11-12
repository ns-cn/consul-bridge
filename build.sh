#!/bin/zsh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o ./releases/consul_bridge_linux .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-w -s" -o ./releases/consul_bridge_win.exe .
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s" -o ./releases/consul_bridge_darwin .