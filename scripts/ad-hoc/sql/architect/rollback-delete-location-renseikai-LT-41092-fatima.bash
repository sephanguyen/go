#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="fatima"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF


update student_package_class set deleted_at = null
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS')  and resource_path = '-2147483645' and  deleted_at = '2023-04-04 04:04:44.444 +0700';

update student_package_access_path set deleted_at = null
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS')  and resource_path = '-2147483645' and  deleted_at = '2023-04-04 04:04:44.444 +0700';

update student_packages set deleted_at = nullwhere 
location_ids && '{01GP3BN89ES8VHECCVRCHDDV64,01GP3CMSRY7T288H60GEGDVEWS}'::TEXT[] and resource_path='-2147483645' and  deleted_at = '2023-04-04 04:04:44.444 +0700';

EOF

