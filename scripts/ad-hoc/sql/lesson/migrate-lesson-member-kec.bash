#!/bin/bash

set -euo pipefail
DB_NAME="lessonmgmt"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
#download data
wget https://storage.googleapis.com/stag-manabie-backend/lessonmgmt-upload/lessonmember.tar.gz
mkdir -p kec-migration && tar xf lessonmember.tar.gz -C kec-migration
FILE_PATH="$PWD/kec-migration/lesson_member.csv"

#import data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy lesson_members(lesson_id,user_id,course_id,updated_at,created_at,resource_path,attendance_status,attendance_reason) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."