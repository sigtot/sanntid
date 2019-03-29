#!/bin/sh

service ssh start

export GOROOT=/usr/lib/go
export GOPATH=/go
su elev -c "tmux new-session -d -s sesh './simelevserver'"
su elev -c "tmux splitw -h -p 66 -d -t sesh 'GOPATH=/root/go go run /root/go/src/github.com/sigtot/sanntid/main.go'"
tail -f /dev/null
