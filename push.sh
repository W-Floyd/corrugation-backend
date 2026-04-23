#!/bin/bash

# sudo docker build -t w-floyd/corrugation . || {
docker buildx build --platform linux/amd64 -t w-floyd/corrugation . || {
    echo 'Failure to build'
    exit
}

docker save w-floyd/corrugation | bzip2 | ssh "${1}" docker load
ssh "${1}" docker compose -f /home/william/conf/vps/docker-compose.yaml --project-directory /home/william/conf/vps up --remove-orphans -d corrugation

exit
