#!/bin/bash

# VARIABLES
TUBED_DIR="/etc/tubed"
TUBED_USER="tubed"
TUBED_TOKEN=${1}
#TUBED_SERVICE="tubed.service"

# CHECK IF RUN AS ROOT
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root" 
   exit 1
fi

# CREATE TUBED USER
#useradd -r -s /bin/false ${TUBED_USER}

# CREATE TUBED DIR
mkdir -p ${TUBED_DIR}

# CREATE TOKEN FILE
echo ${TUBED_TOKEN} > ${TUBED_DIR}/token
chmod 400 ${TUBED_DIR}/token

# CHANGE DIRECTORY/FILES OWNER
#chown -R ${TUBED_USER}:${TUBED_USER} ${TUBED_DIR}

# GET TUBED BINARY
# INSTALL TUBED BINARY
cp tubed /usr/bin/tubed
chmod 755 /usr/bin/tubed

# INSTALL SERVICE FILE
#cp tubed.service /etc/systemd/system/
#chmod 664 /etc/systemd/system/tubed.service