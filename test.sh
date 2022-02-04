#!/bin/bash

__username="$(yq -r '.username' <"${HOME}/.corrugation-backend.yaml")"
__password="$(yq -r '.password' <"${HOME}/.corrugation-backend.yaml")"

__jwt=$(curl -s -X POST -d "username=${__username}" -d "password=${__password}" localhost:8083/login | jq -r '.token')

__curl() {
    __path="${1}"
    shift
    curl "localhost:8083${__path}" "${@}" -H "Authorization: Bearer ${__jwt}"

    echo
}

# __curl '/api/info'

__filename="$(__curl '/api/artifact/upload' -F 'file=@./test.png')"

__curl '/api/artifact/list' -X GET | jq 

# __curl "/api/artifact/get/${__filename}" -X GET -o ignore/download.png

__location_id="$(__curl '/api/location' -X POST -F 'name=Test Location')"

__curl "/api/location/${__location_id}" -X GET | jq

__curl '/api/location/list' -X GET | jq 

__curl "/api/location/${__location_id}/qrcode" -X GET -o ignore/qr.png

# __curl "/api/location/${__location_id}" -X DELETE

# echo

exit
