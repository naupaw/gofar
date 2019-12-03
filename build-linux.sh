#!/bin/bash
rm -rf build
mkdir build
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/gofar github.com/pedox/gofar/server
cp schema.yaml build/
