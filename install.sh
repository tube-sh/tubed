#!/bin/bash

# VARIABLES
TUBED_DIR="/etc/tubed"
#TUBED_USER="tubed"
TUBED_TOKEN=${1}
#TUBED_SERVICE="tubed.service"

# CHECK IF RUN AS ROOT
if [[ $EUID -ne 0 ]]; then
   echo "script must be run as root" 
   exit 1
fi

# CREATE TUBED USER
#useradd -r -s /bin/false ${TUBED_USER}

# CREATE TUBED DIR
echo "create ${TUBED_DIR} directory"
mkdir -p ${TUBED_DIR}

# CREATE TOKEN FILE
echo "write token in ${TUBED_DIR}/token file"
echo ${TUBED_TOKEN} > ${TUBED_DIR}/token

echo "set rights to ${TUBED_DIR}/token file"
chmod 400 ${TUBED_DIR}/token

# CHANGE DIRECTORY/FILES OWNER
#chown -R ${TUBED_USER}:${TUBED_USER} ${TUBED_DIR}

# GET TUBED BINARY
echo "get tubed archive and extract it"
wget -qO- https://github.com/tube-sh/tubed/releases/download/v0.0.1-alpha/tubed_v0.0.1-alpha_linux_amd64.tar.gz | tar -xvz -C /usr/bin/

# INSTALL TUBED BINARY
#cp tubed /usr/bin/tubed
#chmod 755 /usr/bin/tubed

# INSTALL SERVICE FILE
#cp tubed.service /etc/systemd/system/
#chmod 664 /etc/systemd/system/tubed.service

