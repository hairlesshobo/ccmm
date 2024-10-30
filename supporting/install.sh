#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SCRIPT_DIR_SAFE=$(echo -n "$SCRIPT_DIR" | sed 's/\//\\\//g')

GIM_PATH=$(readlink -f $SCRIPT_DIR/gim)
OWNER=$(stat --printf="%U" $GIM_PATH)
GROUP=$(stat --printf="%G" $GIM_PATH)
SED_EXPRESSION="s/__INSTALL_PATH__/$SCRIPT_DIR_SAFE/g; s/__INSTALL_USER__/$OWNER/g; s/__INSTALL_GROUP__/$GROUP/g"

## patch and copy the udev rule
echo "Installing udev rule and reloading udev rules..."
sed "$SED_EXPRESSION" $SCRIPT_DIR/supporting/99-gim.rules > /etc/udev/rules.d/99-gim.rules
udevadm control --reload-rules

## patch and copy the polkit rules, if needed
if [ -d /etc/polkit-1/rules.d ]; then
    echo "Installing polkit rule..."
    sed "$SED_EXPRESSION" $SCRIPT_DIR/supporting/99-gim-policy.rules > /etc/polkit-1/rules.d/99-gim-policy.rules
fi

## patch and copy the systemd service
echo "Creating and starting systemd service..."
sed "$SED_EXPRESSION" $SCRIPT_DIR/supporting/gim.service > /etc/systemd/system/gim.service
systemctl daemon-reload
systemctl enable gim.service
systemctl restart gim.service
