#!/bin/bash

set -euo pipefail

DB_NAME="lessonmgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
BEGIN;
    do
    $$
    begin
        create temp table tmp_partner_dynamic_form_field_values as 
   	 	select * from partner_dynamic_form_field_values where resource_path != '-2147483642';
   	    truncate partner_dynamic_form_field_values;
   	    insert into partner_dynamic_form_field_values
   	    table tmp_partner_dynamic_form_field_values;
    end;
    $$;
COMMIT;
EOF
