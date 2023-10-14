#!/bin/bash

# This script migrate export btn for AIC && GA

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF


UPDATE internal_configuration_value 
SET config_value = '["class","location","locationType","course","courseType","grades","grade","subject","courseAccessPath","schoolLevel","school","schoolCourse","schoolLevelGrade","userTag","notificationTag"]'
WHERE config_key = 'arch.master_management.enable_export' and resource_path in ('-2147483645');



EOF

