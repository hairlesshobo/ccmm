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
SCRIPT_DIR_SAFE=$(echo -n "$SCRIPT_DIR" | sed 's/\//\\\//g')

IMPORTER_PATH=$(readlink -f $SCRIPT_DIR/importer)
OWNER=$(stat --printf="%U" $IMPORTER_PATH)
GROUP=$(stat --printf="%G" $IMPORTER_PATH)
SED_EXPRESSION="s/__INSTALL_PATH__/$SCRIPT_DIR_SAFE/g; s/__INSTALL_USER__/$OWNER/g; s/__INSTALL_GROUP__/$GROUP/g"

if [ ! -f "$SCRIPT_DIR/config.yml" ]; then
    echo "Creating config file..."
    cp "$SCRIPT_DIR/config.example.yml" "$SCRIPT_DIR/config.yml"
fi

## patch and copy the udev rule
echo "Installing udev rule and reloading udev rules..."
sed "$SED_EXPRESSION" $SCRIPT_DIR/supporting/99-ccmm_importer.rules > /etc/udev/rules.d/99-ccmm_importer.rules
udevadm control --reload-rules

## patch and copy the polkit rules, if needed
if [ -d /etc/polkit-1/rules.d ]; then
    echo "Installing polkit rule..."
    sed "$SED_EXPRESSION" $SCRIPT_DIR/supporting/99-ccmm_importer-policy.rules > /etc/polkit-1/rules.d/99-ccmm_importer-policy.rules
fi

## patch and copy the systemd service
echo "Creating and starting systemd service..."
sed "$SED_EXPRESSION" $SCRIPT_DIR/supporting/ccmm_importer.service > /etc/systemd/system/ccmm_importer.service
systemctl daemon-reload
systemctl enable ccmm_importer.service
systemctl restart ccmm_importer.service
