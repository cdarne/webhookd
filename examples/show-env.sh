#!/bin/bash

logfile=./webhook.log
# set stdout to the $logfile
exec 1>$logfile

echo "========================="
echo "Got a Webhook:"
while read line
do
  echo "$line"
done

echo "--------------------------"
env
