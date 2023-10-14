#!/bin/bash

set -euo pipefail
DB_NAME="calendar"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
#download data
wget https://storage.googleapis.com/stag-manabie-backend/lessonmgmt-upload/scheduler.tar.gz
mkdir -p kec-migration && tar xf scheduler.tar.gz -C kec-migration
FILE_PATH="$PWD/kec-migration/scheduler.csv"

#import data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy scheduler(scheduler_id,start_date,end_date,freq,created_at,updated_at,resource_path) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."