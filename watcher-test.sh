#!/bin/sh

while true; do
    sleep 1s
    ./test.sh
    inotifywait corrugation-backend
done
