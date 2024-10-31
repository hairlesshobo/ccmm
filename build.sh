#!/bin/bash
# =================================================================================
#
#		gim - https://www.foxhollow.cc/projects/gim/
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

# for local dev purposes, protect the config file
if [ -f ./dist/config.yml ]; then
    mv ./dist/config.yml ./config.dist.safe.yml
fi

rm -fr ./dist
mkdir -p ./dist/supporting
go build -o ./dist gim.go
cp -r ./supporting/* ./dist/supporting/
mv ./dist/supporting/install.sh ./dist/
cp ./config.yml ./dist/config.example.yml

if [ -f ./config.dist.safe.yml ]; then
    mv ./config.dist.safe.yml ./dist/config.yml
fi