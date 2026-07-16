#!/bin/sh
cd /home/liguanghua/sub2api
docker build -t sub2api-kiro:local \
  --build-arg GOPROXY=https://goproxy.cn,direct \
  --build-arg GOSUMDB=sum.golang.google.cn \
  -f /home/liguanghua/sub2api/Dockerfile /home/liguanghua/sub2api
echo "BUILD_RC=$?"