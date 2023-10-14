#!/bin/bash

set -euo pipefail
DB_NAME="fatima"

# install wget
apt-get update -y
apt-get install -y wget

SECONDS=0
# download CSV from Google Cloud Storage
wget https://storage.googleapis.com/stag-manabie-backend/payment-uat-data-migration/order.csv
mkdir -p payment-migration && cp order.csv payment-migration
FILE_PATH="$PWD/payment-migration/order.csv"

# import the data
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
      -c "BEGIN;" \
      -c "ALTER TABLE public.order DISABLE TRIGGER fill_in_order_seq;" \
      -c "\copy public.order(order_id,student_id,location_id,order_comment,order_status,created_at,updated_at,resource_path,order_type,student_full_name,is_reviewed,withdrawal_effective_date,reason,background,future_measures,loa_start_date,loa_end_date,version_number,order_sequence_number) from "$FILE_PATH" DELIMITER ',' csv header;"\
      -c "ALTER TABLE public.order ENABLE TRIGGER fill_in_order_seq;" \
      -c "COMMIT;"

duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."
