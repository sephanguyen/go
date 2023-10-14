#!/bin/bash

set -euo pipefail

DB_NAME="bob"

USER_ID=$1
RESOURCE_PATH=$2

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, is_system, resource_path)
VALUES('${USER_ID}', 'COUNTRY_JP', 'OpenAPI', '', NULL, 'openapi+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, true, '${RESOURCE_PATH}')
ON CONFLICT DO NOTHING;
EOF
