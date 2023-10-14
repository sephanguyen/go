#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE users
SET previous_name = concat(name,';',last_name,';',first_name),
name = last_name,
last_name = regexp_replace(split_part(last_name,first_name,1),'[[:space:]]|　','')
WHERE (last_name like concat('%　',first_name) or last_name like concat('% ',first_name))
AND name = concat(last_name,' ',first_name)
AND deleted_at IS NULL
AND resource_path = ANY('{${ORG_ID}}');
EOF

### Description ###
# split_part(last_name,first_name,1) mean split last_name by first_name and get first part
# regexp_replace(split_part(last_name,first_name,1),'[[:space:]]|　','') mean after splitting we remove every white space

### Rollback script ###
#previous_name = concat(name,';',last_name,';',first_name)

### Example ###

# Before migrate
#  | name         | last_name | first_name |
#  | John Doe Doe | John Doe  | Doe        | 

# After migrate
#  | name     | last_name | first_name |
#  | John Doe | John      | Doe        |
