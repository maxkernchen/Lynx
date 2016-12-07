#!/usr/bin/env bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <Lynk Name>"
    exit -1
fi

# Copy Meta To Agora
scp ~/Lynx/$1/meta.info mfkernchen1@agora.cs.wcu.edu:~/capPresentation

# Copies Meta From Agora To Anders & Sisko
scp mfkernchen1@agora.cs.wcu.edu:~/capPresentation/meta.info mfkernchen1@sisko.ddns.wcu.edu:
scp mfkernchen1@agora.cs.wcu.edu:~/capPresentation/meta.info mfkernchen1@anders.ddns.wcu.edu:

echo "meta.info Successfully Copied"
exit
