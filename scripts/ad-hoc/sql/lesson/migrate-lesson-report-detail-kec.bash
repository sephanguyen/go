#!/bin/bash

set -euo pipefail
DB_NAME="lessonmgmt"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
#download data
wget https://storage.googleapis.com/stag-manabie-backend/lessonmgmt-upload/lesson_reports_detail.tar.gz
mkdir -p kec-migration && tar xf lesson_reports_detail.tar.gz -C kec-migration
FILE_PATH="$PWD/kec-migration/lesson_reports_detail.csv"

#import data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy lesson_report_details(lesson_report_id,student_id,created_at,updated_at,resource_path,lesson_report_detail_id,report_version) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."