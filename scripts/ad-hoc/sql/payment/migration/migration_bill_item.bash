#!/bin/bash

set -euo pipefail
DB_NAME="fatima"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
# download CSV from Google Cloud Storage
wget https://storage.googleapis.com/stag-manabie-backend/payment-uat-data-migration/bill_item.csv
mkdir -p payment-migration && cp bill_item.csv payment-migration
FILE_PATH="$PWD/payment-migration/bill_item.csv"

# import the data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "BEGIN;" \
      -c "ALTER TABLE bill_item DISABLE TRIGGER fill_in_bill_item_seq;" \
      -c "\copy bill_item(product_id,product_description,product_pricing,discount_amount_type,discount_amount_value,tax_id,tax_category,tax_percentage,created_at,updated_at,resource_path,order_id,bill_type,billing_status,billing_date,discount_amount,tax_amount,final_price,student_id,student_product_id,billing_approval_status,billing_item_description,location_id,location_name,discount_id,adjustment_price,is_latest_bill_item,is_reviewed,raw_discount_amount,bill_item_sequence_number,reference) from "$FILE_PATH" DELIMITER ',' csv header;"\
      -c "ALTER TABLE bill_item ENABLE TRIGGER fill_in_bill_item_seq;" \
      -c "COMMIT;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."
