#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update organizations set 
  scrypt_signer_key = 'm1bxavAKVGLjaG1r+dXdVfYODwuAhJedmPg1oke70S+WBeM+U8lOHiUWnk7LaiPX5LK/B7K1BxoLA5KN5ahCoUI1oeSvOWZbJ8itSZ+4nwqvwJOJNeQ3o769On5wUQoP',
  scrypt_salt_separator = 'k2p4CpgU/wu+mmM0grX0aA==', 
  scrypt_rounds = '/t1Nx1PCgtxneVO9fga7FA==',
  scrypt_memory_cost = 'JFv9JI6mKr4E94gcXkb4hg=='
where organization_id in ('-2147483641');

update organizations set 
  scrypt_signer_key = 'AHH76+4r+5NDnh2aI9UBkvUWvtGfvqaFsM7PLefXNFy7jm904KAGF+Eg8W2d8VSM8tWr5vCeUArhgNbRSnaOtLId9h9Q/BRSoCspCpMDi6AwLk+EMzZnrdybs8JxNipy',
  scrypt_salt_separator = 'k2p4CpgU/wu+mmM0grX0aA==', 
  scrypt_rounds = '/t1Nx1PCgtxneVO9fga7FA==',
  scrypt_memory_cost = 'JFv9JI6mKr4E94gcXkb4hg=='
where organization_id in ('-2147483644');


EOF
