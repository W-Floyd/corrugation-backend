#!/bin/sh

while true; do
    go build
    ./corrugation-backend &
    PID=$!
    inotifywait -r -e modify --include '.*\.go' .
    kill $PID
done
