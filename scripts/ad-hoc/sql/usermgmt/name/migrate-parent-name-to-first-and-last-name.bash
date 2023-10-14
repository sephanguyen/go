set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE users AS u
SET last_name = substring(trim(u.name), '[^\s]+'),
first_name = CASE
WHEN array_length(regexp_split_to_array(trim(u.name), ' +|　+'),1) > 1
THEN trim(regexp_replace(trim(u.name), '.*?\s', ''))
ELSE ''
END
FROM parents as p
WHERE u.name != '' AND u.last_name = '' AND u.first_name = '' AND p.parent_id = u.user_id
AND u.deleted_at IS NULL
AND u.resource_path = ANY('{${ORG_ID}}');
EOF
