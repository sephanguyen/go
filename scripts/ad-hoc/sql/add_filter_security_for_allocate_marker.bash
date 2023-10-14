#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="eureka"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
BEGIN;
    do
    $$
    declare
    resource_path text[]= '{-2147483648,-2147483647,-2147483646,-2147483645,-2147483644,-2147483643,-2147483642,-2147483641,-2147483640,-2147483639,-2147483638,-2147483637,-2147483635,-2147483634,-2147483631,-2147483630,-2147483629}';
    role_names text[] = '{School Admin,HQ Staff,Centre Manager,Teacher,Teacher Lead}';
    _rs_path text;
    _role_name text;
    _read_permission_id text;
    _role_id text;
    begin
        FOREACH _rs_path IN ARRAY resource_path
        LOOP 
            _read_permission_id := generate_ulid();
            INSERT INTO public.permission (permission_id, permission_name, created_at, updated_at, resource_path)
            VALUES 
            (_read_permission_id, 'syllabus.allocate_marker.read', now(), now(), _rs_path)
            ON CONFLICT DO NOTHING;
            FOREACH _role_name IN ARRAY role_names
            LOOP 
                _role_id := (select role_id from role r where r.role_name = _role_name and r.resource_path = _rs_path and role_id is not null limit 1);
                INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
                VALUES 
                (_read_permission_id, coalesce(_role_id,''), now(), now(), _rs_path)
                ON CONFLICT DO NOTHING;
            END LOOP;
        END LOOP;
    end;
    $$;

    do
    $$
    declare
    resource_path text[]= '{-2147483648,-2147483647,-2147483646,-2147483645,-2147483644,-2147483643,-2147483642,-2147483641,-2147483640,-2147483639,-2147483638,-2147483637,-2147483635,-2147483634,-2147483631,-2147483630,-2147483629}';
    role_names text[] = '{School Admin,HQ Staff,Centre Manager}';
    _rs_path text;
    _role_name text;
    _write_permission_id text;
    _role_id text;
    begin
        FOREACH _rs_path IN ARRAY resource_path
        LOOP 
            _write_permission_id := generate_ulid();
            INSERT INTO public.permission (permission_id, permission_name, created_at, updated_at, resource_path)
            VALUES 
            (_write_permission_id, 'syllabus.allocate_marker.write', now(), now(), _rs_path)
            ON CONFLICT DO NOTHING;
            FOREACH _role_name IN ARRAY role_names
            LOOP 
                _role_id := (select role_id from role r where r.role_name = _role_name and r.resource_path = _rs_path and role_id is not null limit 1);
                INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
                VALUES 
                (_write_permission_id, coalesce(_role_id,''), now(), now(), _rs_path)
                ON CONFLICT DO NOTHING;
            END LOOP;
        END LOOP;
    end;
    $$;
COMMIT;
EOF
