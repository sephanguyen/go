#!/bin/bash

set -euo pipefail
DB_NAME="lessonmgmt"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
#download data
wget https://storage.googleapis.com/stag-manabie-backend/lessonmgmt-upload/lesson_reports.tar.gz
mkdir -p kec-migration && tar xf lesson_reports.tar.gz -C kec-migration
FILE_PATH="$PWD/kec-migration/lesson_reports.csv"

#import data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy lesson_reports(lesson_report_id,report_submitting_status,created_at,updated_at,resource_path,form_config_id,lesson_id) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."