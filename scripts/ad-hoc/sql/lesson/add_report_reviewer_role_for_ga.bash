#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
-- Create role: Report Reviewer for GA only
--- Add role ---
INSERT INTO public.role 
  (role_id, role_name, created_at, updated_at, resource_path, is_system)
VALUES 
  ('01GS7JN8KCQ5MC185K88311ZHP', 'Report Reviewer', now(), now(), '-2147483644', true), -- GA
  ('01GS7JN8KCQ5MC185K88311ZHQ', 'Report Reviewer', now(), now(), '-2147483648', true) -- GA Test
ON CONFLICT DO NOTHING;

--- Add lesson.report.review permission ---
INSERT INTO public.permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GS7JJ7W4ATHFT6YT4M4CBK9E', '01GS7JN8KCQ5MC185K88311ZHP', now(), now(), '-2147483644'),
  ('01GS7JJ7W4ATHFT6YT4M4CBK9A', '01GS7JN8KCQ5MC185K88311ZHQ', now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;

--- Add User Group --- 
INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GS7JN8KE5SBA4XCZKBM8689H', 'Report Reviewer', false, now(), now(), '-2147483644'),
  ('01GS7JN8KE5SBA4XCZKBM8689I', 'Report Reviewer', false, now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;

--- Grant role to User group ---
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GS7JNA7E3NGMZ9R48QTX3G9D', '01GS7JN8KE5SBA4XCZKBM8689H', '01GS7JN8KCQ5MC185K88311ZHP', now(), now(), '-2147483644'),
  ('01GS7JNA7E3NGMZ9R48QTX3G9E', '01GS7JN8KE5SBA4XCZKBM8689I', '01GS7JN8KCQ5MC185K88311ZHQ', now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;

--- Grant location to a role ---
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GS7JNA7E3NGMZ9R48QTX3G9D', '01FR4M51XJY9E77GSN4QZ1Q9N6', now(), now(), '-2147483644'),
  ('01GS7JNA7E3NGMZ9R48QTX3G9E', '01FR4M51XJY9E77GSN4QZ1Q9N1', now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;

--- Upsert granted_permission ---
INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
    SELECT * FROM retrieve_src_granted_permission('01GS7JN8KE5SBA4XCZKBM8689H')
    UNION
    SELECT * FROM retrieve_src_granted_permission('01GS7JN8KE5SBA4XCZKBM8689I')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;
EOF
