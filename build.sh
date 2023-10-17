#!/bin/sh
name=pengui

rm -rf build
mkdir -p build

cd build
env GOOS=js GOARCH=wasm go build -o $name.wasm ..
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
cp ../index.html .
cp -r ../assets .

[ "$1" = "serve" ] && gopherjs serve

