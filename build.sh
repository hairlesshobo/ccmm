#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd $SCRIPT_DIR

rm -fr ./dist
mkdir -p ./dist/supporting
go build -o ./dist gim.go
cp -r ./supporting/* ./dist/supporting/
mv ./dist/supporting/install.sh ./dist/
cp ./config.yml ./dist/
