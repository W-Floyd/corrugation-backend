#!/bin/bash

__username="$(yq -r '.username' <"${HOME}/.corrugation-backend.yaml")"
__password="$(yq -r '.password' <"${HOME}/.corrugation-backend.yaml")"

__jwt=$(curl -sS -X POST -d "username=${__username}" -d "password=${__password}" localhost:8083/login | jq -r '.token')

__curl() {
    __path="${1}"
    shift
    curl -sS "localhost:8083${__path}" "${@}" -H "Authorization: Bearer ${__jwt}" | jq

    echo
}

__make_entity() {
    __curl '/api/entity' -X POST -F 'name=Test entity' | jq -r '.'
}

__locate_entity() {
    __curl "/api/entity/${1}" -X PATCH -H 'Content-Type: application/json' -d '{"location":'"${2}"'}'
}

__eat() {
    cat >/dev/null
}

__curl '/api/reset'

# __curl '/api/info'

# __filename="$(__curl '/api/artifact/upload' -F 'file=@./test.png')"

# __curl '/api/artifact/list' -X GET | jq

# __curl "/api/artifact/get/${__filename}" -X GET -o ignore/download.png

__n='7'
echo "Making ${__n} entities"
for n in $(seq 1 "${__n}"); do
    __make_entity | __eat
done

echo 'Listing the entities'
__curl '/api/entity/list' -X GET

echo 'Reading the last entity'
__entity_id="${__n}"
__curl "/api/entity/${__entity_id}" -X GET

# __curl "/api/entity/${__entity_id}/qrcode" -X GET -o ignore/qr.png

echo 'Updating using description (patching)'
__curl "/api/entity/${__entity_id}" -X PATCH -H 'Content-Type: application/json' -d '{"description":"Test desc"}'

echo 'Updating using entity name (replacing)'
__curl "/api/entity/${__entity_id}" -X PUT -H 'Content-Type: application/json' -d '{"name":"Updated name"}'

echo 'Making all but last entity located in last entity'
seq 0 "$((__n - 1))" | while read -r __val; do
    __locate_entity "${__val}" "${__n}" | __eat
done

# echo 'Listing entities located in last entity'
# __curl "/api/entity/${__entity_id}/contains" -X GET

# echo 'Making all entities nested in last entity'
# seq 0 "$((__n - 1))" | while read -r __val; do
#     __locate_entity "${__val}" "$((__val + 1))" | __eat
# done

# echo 'Listing entities located directly in last entity'
# __curl "/api/entity/${__entity_id}/contains" -X GET

# echo 'Listing entities located recursively in last entity'
# __curl "/api/entity/${__entity_id}/contains" -X GET -F 'recursive=true'

seq 1 "${__n}" | while read -r __val; do
    __curl "/api/entity/${__val}" -X PATCH -H 'Content-Type: application/json' -d '{"description":"A test entry, this is entry ID '${__val}'"}'
done | __eat

seq "$((__n / 2))" "${__n}" | while read -r __val; do
    __curl "/api/entity/${__val}" -X PATCH -H 'Content-Type: application/json' -d '{"artifacts":[1]}'
done | __eat

__curl "/api/entity/find/children/0/full" -X GET
__curl "/api/entity/find/locations/full" -X GET

# __curl "/api/entity/${__entity_id}" -X DELETE

# echo

exit
