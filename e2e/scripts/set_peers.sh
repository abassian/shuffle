#!/bin/bash

# obtain the current peers.json from a randomly selected running peer, and copy
# it to the node's config folder. 

mydir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

PEERSFILE=/tmp/peers.json.$$

$mydir/get_current_peers.sh $PEERSFILE

if [ ! -f $PEERSFILE ] ; then
  (>&2 echo "Error getting Peers File. Aborting.")
  exit 1
fi

for node in "$@"
do
   docker cp $PEERSFILE $node:/home/1000/.shuffle/huron/peers.json
done

rm -f $PEERSFILE
