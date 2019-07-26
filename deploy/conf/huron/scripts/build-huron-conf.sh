#!/bin/bash

# This script creates the configuration for a Huron testnet with a variable
# number of nodes.
#
# The Huron validator keys are the same as the prefunded genesis-account keys.
# This script reuses the keys in conf/eth which were generated in a previous
# step by build-eth-conf.sh.
#
# This script also generates a peers.json file in the format used by Huron. The
# files are copied into individual folders which can be used as the datadir for
# each Huron instance.

set -e

N=${1:-4}
IPBASE=${2:-node}
IPADD=${3:-0}
DEST=${4:-"$PWD/conf"}
PORT=${5:-1337}
VALIDATORS=${6:-4}

# Simple sanity check, no more validators than nodes
if [ "$N" -lt "$VALIDATORS" ] ; then
	VALIDATORS="$N"
fi


l=$((N-1))
v=$((VALIDATORS-1))

# use 'shl keys inspect' to extract the raw keys from the Ethereum-style
# keystore files.
for i in $(seq 0 $l)
do
	huron_dest=$DEST/node$i/huron
	eth_source=$DEST/node$i/eth

	mkdir -p $huron_dest
	echo "Generating key pair for node$i"

	docker run \
		-u $(id -u) \
		-v $eth_source:/.shuffle \
		--rm \
		abassian/shuffle:latest keys --json --passfile /.shuffle/pwd.txt inspect --private /.shuffle/keystore/key.json  \
		> $huron_dest/key_info

	awk '/PublicKey/ { gsub("[\",]", ""); print $2 }' $huron_dest/key_info >> $huron_dest/key.pub
	awk '/PrivateKey/ { gsub("[\",]", ""); print $2 }' $huron_dest/key_info >> $huron_dest/priv_key

	echo "$IPBASE$(($IPADD + $i)):$PORT" >> $huron_dest/addr
done

# create the peers.json file
PFILE=$DEST/peers.full.json
GFILE=$DEST/peers.genesis.json
echo "[" > $PFILE
echo "[" > $GFILE

for i in $(seq 0 $l)
do
	dest=$DEST/node$i/huron

	com=","
        com2=","
	if [[ $i == $l ]]; then
		com=""
	fi
	if [[ $i == $v ]]; then
		com2=""
	fi

	printf "\t{\n" >> $PFILE
	printf "\t\t\"NetAddr\":\"$(cat $dest/addr)\",\n" >> $PFILE
	printf "\t\t\"PubKeyHex\":\"0x$(cat $dest/key.pub)\"\n" >> $PFILE
	printf "\t}%s\n"  $com >> $PFILE


	if [ $i -le $v ]; then
		printf "\t{\n" >> $GFILE
		printf "\t\t\"NetAddr\":\"$(cat $dest/addr)\",\n" >> $GFILE
		printf "\t\t\"PubKeyHex\":\"0x$(cat $dest/key.pub)\"\n" >> $GFILE
		printf "\t}%s\n"  $com2 >> $GFILE
	fi
done
echo "]" >> $PFILE
echo "]" >> $GFILE

cp $DEST/peers.genesis.json $DEST/peers.json
for i in $(seq 0 $l)
do
	dest=$DEST/node$i/huron
	cp $DEST/peers.genesis.json $dest/
	cp $DEST/peers.genesis.json $dest/peers.json
done
