#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update organizations set 
  tenant_id = 'prod-manabie-bj1ok',
  scrypt_signer_key = 'kzX6tr+dlpRqj+AYy1GO3SnvN4zLusrq0s4qnI2fh5sebtQhPYlItTWT6GbjV7BRR9fIPfbWWruZtIgyM1k+vP6mabY0/oPfJhLxMboQVudFeT6fa92dTmRYLMOJFx9u',
  scrypt_salt_separator = 'gFa5nh6+ixPSO1xUVZX1UQ==', 
  scrypt_rounds = 'pIemHiO6+clv9n4Lb5xW/Q==',
  scrypt_memory_cost = '2tQDCj6NZtSdenzy8o0YyA=='
where organization_id = '-2147483648';


update organizations set 
  scrypt_signer_key = 'LRRsUgfY5h8+vR7Xnj7dWC2lndp+0B0NfNTvKGqWIR9AwnlYGAPk3onxNEsobWlYolqumk3LcCTxROEAuEIIRYZNAxV/rhBuYUzRrHaaiKsDmsgBy+xBvZmnwdu0urzO',
  scrypt_salt_separator = 'gFa5nh6+ixPSO1xUVZX1UQ==', 
  scrypt_rounds = 'pIemHiO6+clv9n4Lb5xW/Q==',
  scrypt_memory_cost = '2tQDCj6NZtSdenzy8o0YyA=='
where organization_id in ('-2147483643', '-2147483644');


EOF
