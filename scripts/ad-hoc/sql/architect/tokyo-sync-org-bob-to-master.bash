#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update organizations set 
  scrypt_signer_key = 'YsQHFjDUfweEtDMxja0FWHqDY9VDZb9kM0+yy6b/oAM4s4puAGVAbNEhcwM9azzrHa1Wd13cTpsARM4EiS3CtAHvTrr+TwzhDuUJBp33/8QkKrRTUjxXFuBUQe146pt6',
  scrypt_salt_separator = 'jAYv3UAZ8uDDyITP06xuKg==', 
  scrypt_rounds = 'Rl8JJJPIdHZO/Z9GhMmGaA==',
  scrypt_memory_cost = 'cQlldHuN/SLCKksT07bddg=='
where organization_id = '-2147483646';


EOF
