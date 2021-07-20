#!/bin/bash

__jwt=$(curl -s -X POST -d 'username=jon' -d 'password=shhh!' localhost:8083/login | jq -r '.token')

curl localhost:8083/api/info -H "Authorization: Bearer ${__jwt}"

exit