#!/bin/bash

set -euo pipefail
DB_NAME="fatima"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
# download CSV from Google Cloud Storage
wget https://storage.googleapis.com/stag-manabie-backend/payment-uat-data-migration/student_product.csv
mkdir -p payment-migration && cp student_product.csv payment-migration
FILE_PATH="$PWD/payment-migration/student_product.csv"

# import the data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy student_product(student_product_id,student_id,product_id,upcoming_billing_date,start_date,end_date,product_status,approval_status,updated_at,created_at,deleted_at,resource_path,location_id,updated_from_student_product_id,updated_to_student_product_id,student_product_label,is_unique,root_student_product_id,is_associated,version_number) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."
