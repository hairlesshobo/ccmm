#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
BIN_PATH=""

if [ -f $SCRIPT_DIR/gim ]; then
    BIN_PATH="$SCRIPT_DIR/gim"
elif [ -f $SCRIPT_DIR/../gim ]; then
    BIN_PATH="$SCRIPT_DIR/../gim"
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