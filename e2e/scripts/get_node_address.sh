#!/bin/bash

TMPFILE=/tmp/key.$$

docker cp $1:/home/1000/.shuffle/eth/keystore/key.json $TMPFILE

ADDRESS=$(cat $TMPFILE | cut -c2-53 |  sed -e's/"address"://g;s/"//g')

echo  $ADDRESS

rm -f $TMPFILE

