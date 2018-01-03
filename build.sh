#!/usr/bin/env bash

for ARCH in x86 Alpine;do
    docker build -q -t qnib/$(basename $(pwd)):tmp_${ARCH} -f Dockerfile.${ARCH} .
    docker run --name go-wharfie -d qnib/$(basename $(pwd)):tmp_${ARCH} tail -f /dev/null
    docker export go-wharfie |tar xfz - --include="*go-wharfie"
    mv usr/bin/go-wharfie usr/bin/go-wharfie_${ARCH}
    docker rm -f go-wharfie
    docker rmi -f qnib/$(basename $(pwd)):tmp_${ARCH}
done
go build -o usr/bin/go-wharfie_Darwin
