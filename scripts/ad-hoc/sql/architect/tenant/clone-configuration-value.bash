#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

with internal_mapping (prod_rp, internal_rp) as ( values
('-2147483646','2147483646'),
('-2147483645','2147483645'),
('-2147483643','2147483643'),
('-2147483641','2147483641'),
('-2147483631','2147483631'),
('-2147483630','2147483630'),
('-2147483629','2147483629'),
('-2147483626','2147483626'),
('-2147483625','2147483625'),
('-2147483624','2147483624'))
update internal_configuration_value nv
set config_value = icv.config_value 
	, updated_at = now(), last_editor = 'script_init@manabie.com'
from internal_configuration_value icv 
where icv.config_key= nv.config_key
	and icv.resource_path = (select im.prod_rp from internal_mapping im where im.internal_rp = nv.resource_path);
with internal_mapping (prod_rp, internal_rp) as ( values
('-2147483646','2147483646'),
('-2147483645','2147483645'),
('-2147483643','2147483643'),
('-2147483641','2147483641'),
('-2147483631','2147483631'),
('-2147483630','2147483630'),
('-2147483629','2147483629'),
('-2147483626','2147483626'),
('-2147483625','2147483625'),
('-2147483624','2147483624'))
update external_configuration_value nv
set config_value = ecv.config_value 
	, updated_at = now(), last_editor = 'script_init@manabie.com'
from external_configuration_value ecv 
where ecv.config_key= nv.config_key
	and ecv.resource_path = (select im.prod_rp from internal_mapping im where im.internal_rp = nv.resource_path);

EOF
