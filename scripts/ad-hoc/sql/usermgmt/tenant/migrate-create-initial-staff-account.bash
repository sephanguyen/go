#!/bin/bash

set -euo pipefail

DB_NAME="bob"

USER_ID=$1
USER_EMAIL=$2
USER_FIRST_NAME=$3
USER_LAST_NAME=$4
ORG_ID=$5
USER_GROUP_ID=$6
LOCATION_ID=$7

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
INSERT INTO public.users
(user_id, country, name, first_name, last_name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path)
VALUES('${USER_ID}', 'COUNTRY_JP', '${USER_LAST_NAME} ${USER_FIRST_NAME}', '${USER_FIRST_NAME}', '${USER_LAST_NAME}', '', NULL, '${USER_EMAIL}', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '${ORG_ID}')
ON CONFLICT DO NOTHING;

INSERT INTO public.users_groups
(user_id, group_id, is_origin, status, updated_at, created_at, resource_path)
VALUES('${USER_ID}', 'USER_GROUP_SCHOOL_ADMIN', true, 'USER_GROUP_STATUS_ACTIVE', now(), now(), '${ORG_ID}')
ON CONFLICT DO NOTHING;

INSERT INTO public.school_admins
(school_admin_id, school_id, updated_at, created_at, resource_path)
VALUES('${USER_ID}', ${ORG_ID}, now(), now(), '${ORG_ID}')
ON CONFLICT DO NOTHING;

INSERT INTO public.staff
(staff_id, updated_at, created_at, resource_path)
VALUES('${USER_ID}', now(), now(), '${ORG_ID}')
ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
  user_id,
  user_group_id,
  created_at,
  updated_at,
  resource_path
) VALUES 
('${USER_ID}', '${USER_GROUP_ID}', now(), now(), '${ORG_ID}') ON CONFLICT DO NOTHING;

INSERT INTO user_access_paths(
  user_id,
  location_id,
  created_at,
  updated_at,
  resource_path
) VALUES 
('${USER_ID}', '${LOCATION_ID}', now(), now(), '${ORG_ID}') ON CONFLICT DO NOTHING;
EOF
