#!/bin/bash

__username="$(yq -r '.username' <"${HOME}/.corrugation-backend.yaml")"
__password="$(yq -r '.password' <"${HOME}/.corrugation-backend.yaml")"

__jwt=$(curl -s -X POST -d "username=${__username}" -d "password=${__password}" localhost:8083/login | jq -r '.token')

__curl() {
    __path="${1}"
    shift
    curl "localhost:8083${__path}" "${@}" -H "Authorization: Bearer ${__jwt}"

    echo
    echo
}

# __curl '/api/info'

__curl '/api/upload/box-image' -F 'file=@./screen.png' -F 'box-name=Network Cables'

__curl '/api/list/box-image' -X GET -F 'box-name=Network Cables'

__curl '/api/download/box-image' -X GET -F 'box-name=Network Cables'

exit
