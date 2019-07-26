#!/bin/bash

# This script produces the Ethereum configuration for each node on the testnet.
# It generates a new Ethereum key, controlled by the same password, for each
# node, and aggretates the corresponding public keys in a genesis.json file,
# which is used by shuffle to initialize the accounts in the State.
# Additionally, it creates the base for the shl.toml file used by shuffle to
# read configuration, on top of command-line flags. The configuration files are
# placed in different folders (named node0...nodeN) which can be copied or
# mounted directly in the root directory for shuffle (controlled by 'datadir'
# flag). The output of this script, executed with default parameters, will look
# something like this:
#
#	conf/solo/conf
#	├── genesis.json
#	├── keystore
#	│   ├── node0-key.json
#	│   ├── node1-key.json
#	│   ├── node2-key.json
#	│   └── node3-key.json
#	├── node0
#	│   └── eth
#	│       ├── genesis.json
#	│       ├── keystore
#	│       │   └── key.json
#	│       └── pwd.txt
#	├── node1
#	│   └── eth
#	│       ├── genesis.json
#	│       ├── keystore
#	│       │   └── key.json
#	│       └── pwd.txt
#	├── node2
#	│   └── eth
#	│       ├── genesis.json
#	│       ├── keystore
#	│       │   └── key.json
#	│       └── pwd.txt
#	└── node3
#	    └── eth
#	        ├── genesis.json
#	        ├── keystore
#	        │   └── key.json
#	        └── pwd.txt


#         Invocation line in conf/huron/makefile
#         ./../eth/scripts/build-eth-conf.sh $(NODES) $(DEST) $(PASS) $(VALIDATORS) $(POA)


set -e

N=${1:-4} # number of nodes
DEST=${2:-"$(pwd)/../conf"} # output directory
PASS=${3:-"$(pwd)/../pwd.txt"} # password file for Ethereum accounts
VALIDATORS=${4:-4} # number of validators on genesis whitelist
POA=${5:-true} # is this a POA network?


if [ "$VALIDATORS" -gt "$N" ] ; then # Simple sanity check. We cannot have more nodes than validators
  VALIDATORS=$N
fi

mydir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

l=$((N-1))
v=$((VALIDATORS-1))

for i in $(seq 0 $l)
do
	dest=$DEST/node$i/eth
	mkdir -p $dest
    # Use a Docker container to run the 'shl keys' command that creates
	# accounts. This saves us the trouble of installing shl locally.
	# The file is written directly into the mounted directory.
    docker run --rm \
		-u $(id -u) \
		-v $dest:/datadir \
		-v $PASS:/pwd.txt \
		abassian/shuffle:latest keys --passfile=/pwd.txt generate /datadir/keystore/key.json  | \
    		awk '/Address/ {print $2}'  >> $dest/addr
done

# Generate the genesis file
GFILE=$DEST/genesis.json
echo "{" > $GFILE
printf "\t\"alloc\": {\n" >> $GFILE
for i in $(seq 0 $l)
do
	com=","
	if [[ $i == $l ]]; then
		com=""
	fi
	printf "\t\t\"$(cat $DEST/node$i/eth/addr)\": {\n" >> $GFILE
    printf "\t\t\t\"balance\": \"133700000000000000000$i\",\n" >> $GFILE
    printf "\t\t\t\"moniker\": \"node$i\"\n" >> $GFILE
    printf "\t\t}%s\n" $com >> $GFILE
done
printf "\t}\n" >> $GFILE
echo "}" >> $GFILE



if [ $POA ] ; then
    # Copy the POA contract into place.


	DESTPOA=$DEST/poa
	if [ ! -d "$DESTPOA" ] ; then
	    mkdir "$DESTPOA"
	fi

       cp $mydir/../../../../e2e/smart-contracts/genesis_array.sol $DESTPOA/genesis.sol

    # Generate the pregenesis file
    GFILE=$DESTPOA/pregenesis.json
    echo "{" > $GFILE

    printf "\t\"precompiler\":{\n" >> $GFILE
    printf "\t\t \"contracts\": [\n" >> $GFILE
    printf "\t\t\t {\n" >> $GFILE
    printf "\t\t\t\t \"address\": \"0XABBAABBAABBAABBAABBAABBAABBAABBAABBAABBA\",\n" >> $GFILE
    printf "\t\t\t\t \"filename\": \"genesis.sol\",\n" >> $GFILE
    printf "\t\t\t\t \"authorising\": \"true\",\n" >> $GFILE
    printf "\t\t\t\t \"contractname\": \"POA_Genesis\",\n" >> $GFILE
    printf "\t\t\t\t \"balance\": \"1337000000000000000099\",\n" >> $GFILE
    printf "\t\t\t\t \"preauthorised\": [\n" >> $GFILE


    comma=""

    for i in $(seq 0 $v)
    do
       printf "\t\t\t\t\t $comma{ \"address\": \"$(cat $DEST/node$i/eth/addr)\", \"moniker\": \"node$i\"}\n" >> $GFILE
       comma=","
    done


    printf "\t\t\t\t ]\n" >> $GFILE
    printf "\t\t\t }\n" >> $GFILE
    printf "\t\t ]\n" >> $GFILE
    printf "\t },\n" >> $GFILE



    printf "\t\"alloc\": {\n" >> $GFILE
    for i in $(seq 0 $l)
    do
    	com=","
    	if [[ $i == $l ]]; then
    		com=""
    	fi
    	printf "\t\t\"$(cat $DEST/node$i/eth/addr)\": {\n" >> $GFILE
        printf "\t\t\t\"balance\": \"133700000000000000000$i\",\n" >> $GFILE
        printf "\t\t\t\"moniker\": \"node$i\"\n" >> $GFILE
        printf "\t\t}%s\n" $com >> $GFILE
    done
    printf "\t}\n" >> $GFILE
    echo "}" >> $GFILE
fi



gKeystore=$DEST/keystore
mkdir -p $gKeystore

# Copy files into each node's folder and cleanup
for i in $(seq 0 $l)
do
	dest=$DEST/node$i
	cp $DEST/genesis.json $dest/eth
	cp $PASS $dest/eth
	cp -r $dest/eth/keystore/key.json $gKeystore/node$i-key.json
    rm $dest/eth/addr
done
