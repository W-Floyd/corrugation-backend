#!/bin/sh

while true; do
    go build
    err="${?}"
    if [ "${err}" != '0' ]; then
        echo 'Failed to build...'
        inotifywait -q -r -e modify --include '.*\.go' .
    else
        CORRUGATION_AUTHENTICATION=false ./corrugation-backend &
        PID="${!}"
        inotifywait -q -r -e modify --include '.*\.go' .
        kill "${PID}"
    fi
done
