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
INSERT INTO grade (grade_id,"name",is_archived,partner_internal_id,updated_at,created_at,"sequence",remarks, resource_path)
select generate_ulid(), g."name", g.is_archived, 'test_'||g.partner_internal_id , now(), now(), g."sequence", g.remarks , (select im.internal_rp from internal_mapping im where im.prod_rp= g.resource_path) 
from grade g
where g.resource_path in (select im.prod_rp from internal_mapping im)
	and g.deleted_at is null
	and (select count(*) from grade g2 where g2.resource_path = (select im.internal_rp from internal_mapping im where im.prod_rp= g.resource_path))=0
  
EOF
