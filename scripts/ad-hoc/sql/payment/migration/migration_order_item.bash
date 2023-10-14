#!/bin/bash

set -euo pipefail
DB_NAME="fatima"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
# download CSV from Google Cloud Storage
wget https://storage.googleapis.com/stag-manabie-backend/payment-uat-data-migration/order_item.csv
mkdir -p payment-migration && cp order_item.csv payment-migration
FILE_PATH="$PWD/payment-migration/order_item.csv"

# import the data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "\copy order_item(order_id,product_id,discount_id,start_date,created_at,resource_path,student_product_id,order_item_id,product_name,effective_date,cancellation_date,end_date) from "$FILE_PATH" DELIMITER ',' csv header;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."
