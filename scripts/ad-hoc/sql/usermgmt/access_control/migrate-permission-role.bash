#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1
ROLE_NAME=$2
PERMISSION_NAME=$3

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
with 
role as (
  select r.role_id, r.resource_path
  from "role" r
  where r.role_name = ANY('{${ROLE_NAME}}')
	and r.resource_path = ANY('{${ORG_ID}}')),
permission as (
	select p.permission_id, p.resource_path
  from "permission" p 
  where p.permission_name = ANY('{${PERMISSION_NAME}}')
	and p.resource_path = ANY('{${ORG_ID}}'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;
EOF
