#!/bin/bash
# =================================================================================
#
#		ccmm - https://www.foxhollow.cc/projects/ccmm/
#
#	 go-import-media, aka gim, is a tool for automatically importing media
#	 from removable disks into a predefined folder structure automatically.
#
#		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
#
#		Licensed under the Apache License, Version 2.0 (the "License");
#		you may not use this file except in compliance with the License.
#		You may obtain a copy of the License at
#
#		     http://www.apache.org/licenses/LICENSE-2.0
#
#		Unless required by applicable law or agreed to in writing, software
#		distributed under the License is distributed on an "AS IS" BASIS,
#		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#		See the License for the specific language governing permissions and
#		limitations under the License.
#
# =================================================================================

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd $SCRIPT_DIR

# for local dev purposes, protect the config files
if [ -f ./dist/importer/importer.yml ]; then
    echo "Preserving existing importer config"
    mv ./dist/importer.yml ./config.importer.safe.yml
fi

if [ -d ./dist ]; then
    echo "Removing existing build"
    rm -fr ./dist
fi

echo "Creating new build directories"
mkdir -p ./dist/supporting

echo "Building importer..."
go build -o ./dist/ ./importer/importer.go

echo "Copying supporting files"
cp -r ./supporting/* ./dist/supporting/
mv ./dist/supporting/install.sh ./dist/
cp ./importer/config.yml ./dist/importer.example.yml

if [ -f ./config.importer.safe.yml ]; then
    echo "Recovering existing importer config"
    mv ./config.importer.safe.yml ./dist/importer.yml
fi