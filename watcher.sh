#!/bin/sh

while true; do
    go build
    CORRUGATION_AUTHENTICATION=false ./corrugation-backend &
    PID=$!
    inotifywait -r -e modify --include '.*\.go' .
    kill $PID
done
