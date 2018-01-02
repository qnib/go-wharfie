#!/usr/bin/env bash

docker build -q -t qnib/$(basename $(pwd)) -f Dockerfile.ubuntu .
docker run --name fisherman -d qnib/$(basename $(pwd)) tail -f /dev/null
docker export fisherman|tar xfz - --include="*go-wharfie"
docker rm -f fisherman
for x in $(docker ps -q);do
   docker cp ./usr/bin/go-wharfie ${x}:/usr/bin/
done
