# webhookd, the daemon for Shopify's webhooks 

## TL;DR

`webhookd` is taking care of receiving and verifying your webhooks. All you have to do is providing your program to handle the webhooks.   


## Description

Getting started receiving webhooks can be challenging, because it requires hosting a web server, verifying the webhook and then trigger the business logic. The goal of that project is to provide easy getting started project/templates so that the receiving part of webhooks is already done, and you can focus on the business logic.

Here's an example on how to use it:

Start `webhookd` and tell it which program to run when a webhook is received
```
$ webhookd --shared-secret=abc123 ./my_program.sh
```

Create a script/program to handle the webhooks from STDIN and ENV:
```bash
#!/bin/bash
# my_program.sh

logfile=./webhook.log
# Set stdout to the $logfile
exec 1>$logfile

# Read the webhooks JSON from STDIN and echo it in the $logfile
while read line
do
  echo "$line"
done

# Dump the process ENV variables
env
```

That's all it takes to start processing webhooks! The idea is it to take all the burden of receiving/verifying the webhooks in `webhookd` so that you can focus on the business logic. Any script/program in any language that can read STDIN and ENVS can be used here!

## Main highlights

- It's the UNIX philosophy: write programs that do one thing well and communicate using text streams as a universal interface!
- Yes, it is like CGI but for webhooks!
- it's written in go, so we provide a single exec that is portable on most platforms and  iseasily deployable.
- it might be later a good place to experiment with new transport technologies (HTTP/2, QUIC) and have a better delivery mechanism.

## Credits

This project is heavily inspired by http://websocketd.com/. Thanks for the inspiration and for moving the UNIX philosophy forward <3! 
