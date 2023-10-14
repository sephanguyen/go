#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update internal_configuration_value 
set config_value = '[{"url":"https://www.with-us.co.jp/privacy/","title":"個人情報保護方針"}]'
where config_key = 'urls_widget' and resource_path = '-2147483630';

update internal_configuration_value 
set config_value = '[{"url":"https://managara.nsf-h.ed.jp/privacy","title":"個人情報保護方針"}]'
where config_key = 'urls_widget' and resource_path = '-2147483629';

EOF
