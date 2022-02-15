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
    __curl '/api/entity' -X POST -F "name=${1}" | jq -r '.'
}

__locate_entity() {
    __curl "/api/entity/${1}" -X PATCH -H 'Content-Type: application/json' -d '{"location":'"${2}"'}'
}

__make_locate() {
    __v__="$(__make_entity "${1}")"
    __locate_entity "${__v__}" "${2}" | __eat
    echo "${__v__}"
}

__quantify() {
    __curl "/api/entity/${1}" -X PATCH -H 'Content-Type: application/json' -d '{"metadata": {"quantity":'"${2}"'}}'
}

__describe(){
    __curl "/api/entity/${1}" -X PATCH -H 'Content-Type: application/json' -d '{"description": "'"${2}"'"}'
}

__eat() {
    cat >/dev/null
}

__curl '/api/reset'

__room="$(__make_entity "Room")"

__desk="$(__make_locate "Desk" "${__room}")"
__drawer="$(__make_locate "Drawer" "${__desk}")"
__make_locate "Tape" "${__drawer}"
__make_locate "Pens" "${__drawer}"
__make_locate "Sharpies" "${__drawer}"
__make_locate "Scissors" "${__drawer}"
__make_locate "Ruler" "${__drawer}"
__monitor_stand="$(__make_locate "Monitor Stand" "${__desk}")"
__make_locate "Bose Soundlink Mini II" "${__monitor_stand}"
__make_locate "Bose Soundlink Mini II Charging Cradle" "${__monitor_stand}"
__make_locate "Bose Soundlink Mini II Case" "${__monitor_stand}"
__make_locate "Monitor" "${__monitor_stand}"
__make_locate "DAC" "${__monitor_stand}"
__make_locate "Rubber Duck" "${__monitor_stand}"

__dresser="$(__make_locate "Dresser" "${__room}")"
__dresser_top="$(__make_locate "Top Drawer" "${__dresser}")"
__dresser_bottom="$(__make_locate "Bottom Drawer" "${__dresser}")"
__dresser_middle="$(__make_locate "Middle Drawer" "${__dresser}")"

__v="$(__make_locate "USB Power Supplies" "${__dresser_bottom}")"
__make_locate "Wireless chargers" "${__v}"
__make_locate "Cheap wall chargers" "${__v}"
__make_locate "Fixed micro-USB" "${__v}"

__v="$(__make_locate "Power Cables" "${__dresser_bottom}")"

__usb_cables="$(__make_locate "USB Cables" "${__dresser_bottom}")"

__v="$(__make_locate "Micro USB Cable" "${__usb_cables}")"
# These quantities are made up
__quantify "${__v}" 5

__v="$(__make_locate "Power Supplies" "${__dresser_bottom}")"

__v="$(__make_locate "Computer Parts & Hard Drives" "${__dresser_bottom}")"

__car="$(__make_entity "Car")"
__describe "${__car}" '2005 Nissan Altima'

__car_cabin="$(__make_locate "Cabin" "${__car}")"
__make_locate "Prescription Sunglasses" "${__car_cabin}"
__make_locate "12V USB Charger" "${__car_cabin}"
__make_locate "USB C Cable" "${__car_cabin}"

__car_boot="$(__make_locate "Boot" "${__car}")"
__make_locate "Gas can" "${__car_boot}"
__v="$(__make_locate "Spare tire" "${__car_boot}")"
__make_locate "1/2\" Breaker Bar" "${__v}"
__make_locate "1/4\" hex to 3/8\" square drive adapter" "${__v}"
__make_locate "1/4\" hex to 1/2\" square drive adapter" "${__v}"

__make_locate "Small hex key set, metric" "${__car_boot}"
__make_locate "Small hex key set, inch" "${__car_boot}"

exit
