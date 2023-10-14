#!/bin/bash

set -euo pipefail
DB_NAME="lessonmgmt"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
#download data
wget https://storage.googleapis.com/stag-manabie-backend/lessonmgmt-upload/partner_dynamic_form_field_values_test.tar.gz
mkdir -p kec-migration && tar xf partner_dynamic_form_field_values_test.tar.gz -C kec-migration
FILE_PATH="$PWD/kec-migration/partner_dynamic_form_field_values_test.csv"

#import data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy partner_dynamic_form_field_values(dynamic_form_field_value_id,field_id,lesson_report_detail_id,created_at,updated_at,value_type,string_value,resource_path,int_value) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."