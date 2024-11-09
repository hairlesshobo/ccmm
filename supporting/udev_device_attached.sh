#!/bin/bash
# =================================================================================
#
#		ccmm - https://www.foxhollow.cc/projects/ccmm/
#
#	Connection Church Media Manager, aka ccmm, is a tool for managing all
#   aspects of produced media- initial import from removable media,
#   synchronization with clients and automatic data replication and backup
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
BIN_PATH=""

if [ -f $SCRIPT_DIR/importer ]; then
    BIN_PATH="$SCRIPT_DIR/importer"
elif [ -f $SCRIPT_DIR/../importer ]; then
    BIN_PATH="$SCRIPT_DIR/../importer"
fi

BIN_PATH=$(readlink -f $BIN_PATH)
BIN_DIR=$(dirname $BIN_PATH)
OWNER=$(stat --printf="%U" $BIN_PATH)

echo $BIN_DIR

# udev rules run in parallel, if we move to fast the device won't yet exist in /dev
# hopefully 3 seconds is enough, if not, i'll fix the service to poll a few times
# before giving up
sleep 3
/usr/bin/su - -c "$BIN_PATH device_attached $1" $OWNER