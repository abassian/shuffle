#!/bin/bash

# This script adds Huron configuration to an shuffle shl.toml file. 

set -e

N=${1:-4}
IPBASE=${2:-node}
IPADD=${3:-0}
DEST=${4:-"conf"}

l=$((N-1))

PFILE=$DEST/shl.toml
echo "[huron]" >> $PFILE 
echo "store = true" >> $PFILE
echo "heartbeat = \"50ms\"" >> $PFILE
echo "timeout = \"200ms\"" >> $PFILE
    
for i in $(seq 0 $l) 
do
	dest=$DEST/node$i
	cp $DEST/shl.toml $dest/shl.toml
	echo "listen = \"$IPBASE$(($IPADD +$i)):1337\"" >> $dest/shl.toml
done