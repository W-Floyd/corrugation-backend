#!/bin/bash

sudo docker build -t w-floyd/corrugation-backend .

sudo docker save w-floyd/corrugation-backend | bzip2 | pv | ssh "${1}" docker load
ssh "${1}" docker-compose -f /root/server-config/docker-compose.yml --project-directory /root/server-config up --remove-orphans -d

exit