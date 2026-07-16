#!/bin/sh
cd /home/liguanghua/sub2api
docker run --rm -v /home/liguanghua/sub2api/backend:/src -w /src \
  -e GOPROXY=https://goproxy.cn,direct -e GOSUMDB=sum.golang.google.cn -e GOFLAGS=-mod=mod \
  golang:1.26.3-alpine sh -c 'go test ./internal/service/ -run Kiro -count=1'
echo "TEST_RC=$?"