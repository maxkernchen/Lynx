#!/usr/bin/env bash

# Checks to see if the user supplied 0 or an incorrect # of arguments - it
# tries to recover if the user supplies 0.
if [ $# -ne 3 ]; then
  echo "usage: $0 <initial runtime> <reboot delay> <second runtime>"
  exit 1
fi

cd ../guiserver

# First run of guiserver
go run guiserver.go &
sleep $1 # Sleep for first arg amt
pkill -TERM -P $$

# Reboot of guiserver if 2nd arg > 0
if [ $2 -gt 0 ]; then
  sleep $2 # Sleep for second arg amt
  go run guiserver.go &
  sleep $3 # Sleep for third arg amt
  pkill -TERM -P $$
fi

exit
