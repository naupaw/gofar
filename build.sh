#!/bin/bash
rm -rf build
mkdir build
go build -o build/gofar github.com/pedox/gofar/server
cp schema.yaml build/
