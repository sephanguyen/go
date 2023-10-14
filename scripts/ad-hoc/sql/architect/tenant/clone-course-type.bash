#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

set role postgres;
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
INSERT INTO course_type (course_type_id,"name",created_at,updated_at,resource_path,remarks,is_archived) 
select generate_ulid(), ct."name", now(), now(), (select im.internal_rp from internal_mapping im where im.prod_rp= ct.resource_path), '', false
from course_type ct 
where ct.resource_path in (select im.prod_rp from internal_mapping im)
	and ct.deleted_at is null
	and (select count(*) from course_type ct2 where ct2.resource_path = (select im.internal_rp from internal_mapping im where im.prod_rp= ct.resource_path))=0
  
EOF
