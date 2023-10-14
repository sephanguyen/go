#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update organizations set 
  scrypt_signer_key = 'uRqmtPG6DgorPdH0XYfKSen6XtO3rwhMTnhfT+z26Xq6t9lvZyTZ1FQMfgTIbyIowldgVXc3qRmIdsFcFa1TD2linf78nSXTRdywBQU+XHLRCtP4Z8XJNnHbUYJNti6Y',
  scrypt_salt_separator = 'whAdBQBa516j5kXaqgoOIA==', 
  scrypt_rounds = 'E1H3WoKVaC61n+cNEt22ow==',
  scrypt_memory_cost = 'iTmnwiKXMMlHflnjA06DrQ=='
where organization_id in ('-2147483644');


update organizations set 
  scrypt_signer_key = 'GJMy0GqG0DHiv3KQCOjaweuVXuccsAITWv83zjlzbCvzzaKHIhEBCXd7OuiMgL8O0MkTnUWGXU6E4TFHwUDzOoFTFxLG8NTfrPuOjxOuZjednXrk2NjiAp3BWcjeiUBK',
  scrypt_salt_separator = 'whAdBQBa516j5kXaqgoOIA==', 
  scrypt_rounds = 'E1H3WoKVaC61n+cNEt22ow==',
  scrypt_memory_cost = 'iTmnwiKXMMlHflnjA06DrQ=='
where organization_id in ('-2147483645');

EOF
