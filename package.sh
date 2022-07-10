#!/bin/bash
CLI_VERSION=1.6.1
rm -rf ./bin && mkdir ./bin


# Mac OSX
OUTPUT_PATH=ikatago-$CLI_VERSION-mac-osx
mkdir -p ./bin/$OUTPUT_PATH
GOOS=darwin GOARCH=amd64 go build -o ./bin/$OUTPUT_PATH/ikatago
cd ./bin
zip -r $OUTPUT_PATH.zip $OUTPUT_PATH
cd - >/dev/null

# linux
OUTPUT_PATH=ikatago-$CLI_VERSION-linux
mkdir -p ./bin/$OUTPUT_PATH
GOOS=linux GOARCH=amd64 go build -o ./bin/$OUTPUT_PATH/ikatago
cd ./bin
zip -r $OUTPUT_PATH.zip $OUTPUT_PATH
cd - >/dev/null

# windows 64bit
OUTPUT_PATH=ikatago-$CLI_VERSION-win64
mkdir -p ./bin/$OUTPUT_PATH
GOOS=windows GOARCH=amd64 go build -o ./bin/$OUTPUT_PATH/ikatago.exe
cd ./bin
zip -r $OUTPUT_PATH.zip $OUTPUT_PATH
cd - >/dev/null

# windows 32bit
OUTPUT_PATH=ikatago-$CLI_VERSION-win32
mkdir -p ./bin/$OUTPUT_PATH
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o ./bin/$OUTPUT_PATH/ikatago.exe
cd ./bin
zip -r $OUTPUT_PATH.zip $OUTPUT_PATH
cd - >/dev/null
