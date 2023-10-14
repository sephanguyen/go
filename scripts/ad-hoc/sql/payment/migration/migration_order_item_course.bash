#!/bin/bash

set -euo pipefail
DB_NAME="fatima"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
# download CSV from Google Cloud Storage
wget https://storage.googleapis.com/stag-manabie-backend/payment-uat-data-migration/order_item_course.csv
mkdir -p payment-migration && cp order_item_course.csv payment-migration
FILE_PATH="$PWD/payment-migration/order_item_course.csv"

# import the data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy order_item_course(order_id,package_id,course_id,course_name,course_slot,course_slot_per_week,created_at,updated_at,resource_path,order_item_course_id) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."
