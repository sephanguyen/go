#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE users
SET previous_name = concat(name,';',last_name,';',first_name),
last_name = substring(trim(name), '[^\s]+'),
first_name = CASE
WHEN array_length(regexp_split_to_array(trim(name), '[[:space:]]|　'),1) > 1
THEN trim(regexp_replace(trim(name), '.*?\s', ''))
ELSE ''
END
WHERE name != ''
AND name != trim(concat(last_name,' ',first_name))
AND name != trim(concat(last_name,'　',first_name))
AND deleted_at IS NULL
AND resource_path = ANY('{${ORG_ID}}');
EOF


### Description ###
# substring(trim(name), '[^\s]+') mean get all characters from beginning of the string to the first space

# array_length(regexp_split_to_array(trim(name), '[[:space:]]|　'),1) > 1 mean check the length of name 
# after splitting by space (normal space/japanese space)

# trim(regexp_replace(trim(name), '.*?\s', '')) mean get all characters from the first space to the end


### Rollback script ###
#previous_name = concat(name,';',last_name,';',first_name)

### Example ###

# Before migrate
#  | name             | last_name   | first_name |
#  | John Doe         | John Doe    | Doe        |
#  | Robert Pattinson |             |            |
#  | Tommy Shelby     | TommyShelby |            | 

# After migrate
#  | name             | last_name   | first_name |
#  | John Doe         | John        | Doe        |
#  | Robert Pattinson | Robert      | Pattinson  |
#  | Tommy Shelby     | Tommy       | Shelby     | 
