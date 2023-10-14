#!/bin/bash

# This script truncates data in payment-related tables.

set -euo pipefail

DB_NAME="fatima"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
TRUNCATE public.accounting_category           CASCADE;
TRUNCATE public.bill_item                     CASCADE;
TRUNCATE public.bill_item_course              CASCADE;
TRUNCATE public.billing_ratio                 CASCADE;
TRUNCATE public.billing_schedule              CASCADE;
TRUNCATE public.billing_schedule_period       CASCADE;
TRUNCATE public.discount                      CASCADE;
TRUNCATE public.fee                           CASCADE;
TRUNCATE public.leaving_reason                CASCADE;
TRUNCATE public.material                      CASCADE;
TRUNCATE public.order                         CASCADE;
TRUNCATE public.order_action_log              CASCADE;
TRUNCATE public.order_item                    CASCADE;
TRUNCATE public.order_item_course             CASCADE;
TRUNCATE public.package                       CASCADE;
TRUNCATE public.package_course                CASCADE;
TRUNCATE public.package_course_fee            CASCADE;
TRUNCATE public.package_course_material       CASCADE;
TRUNCATE public.package_quantity_type_mapping CASCADE;
TRUNCATE public.product                       CASCADE;
TRUNCATE public.product_accounting_category   CASCADE;
TRUNCATE public.product_grade                 CASCADE;
TRUNCATE public.product_location              CASCADE;
TRUNCATE public.product_price                 CASCADE;
TRUNCATE public.product_setting               CASCADE;
TRUNCATE public.student_package_by_order      CASCADE;
TRUNCATE public.student_product               CASCADE;
TRUNCATE public.tax                           CASCADE;
EOF
