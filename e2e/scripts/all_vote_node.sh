#!/bin/bash

# for all running nodes, sign and broadcast a vote to accept the nominee

NOMINEENODE=$1

mydir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

RUNNINGNODES=$($mydir/get_running_nodes.sh | awk '{print $1}')
PASSWD=$mydir/../../deploy/conf/eth/pwd.txt
NOMADD=$($mydir/get_node_address.sh $NOMINEENODE)

echo "NOMADD: " $NOMADD

for i in $(echo $RUNNINGNODES)
do			
   $mydir/write_shlc_config.sh $i
   FROMADD=$($mydir/get_node_address.sh $i)
   shlc poa vote $NOMADD --from $FROMADD --verdict true --pwd $PASSWD
done


