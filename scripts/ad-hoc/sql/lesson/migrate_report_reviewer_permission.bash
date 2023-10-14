#!/bin/bash
# This script migrate report review permission for the roles excecpt report reviewer role of GA partner (-2147483643)
# This script must not execute on GA service
set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
  do \$\$
    declare
    resource_path text[]= '{-2147483648,-2147483647,-2147483646,-2147483645,-2147483644,-2147483642,-2147483641,-2147483640,-2147483639,-2147483638,-2147483637,-2147483635,-2147483631,-2147483630,-2147483629,-2147483634,100000,-2147483628,-2147483627,-2147483626,-2147483625,-2147483624,-2147483623}';
    role_names text[] = '{School Admin,HQ Staff,Centre Manager,Centre Lead,Centre Staff,Teacher,Teacher Lead}';
    _rs_path text;
    _role_name text;
    _permission_id text;
    _role_id text;
    begin
        FOREACH _rs_path IN ARRAY resource_path
        LOOP 
            _permission_id := (select permission_id from permission p where p.permission_name = 'lesson.report.review' and p.resource_path = _rs_path and permission_id is not null limit 1);
            IF _permission_id IS NOT NULL THEN        
            FOREACH _role_name IN ARRAY role_names
            LOOP 
                _role_id := (select role_id from role r where r.role_name = _role_name and r.resource_path = _rs_path and role_id is not null limit 1);
                IF _role_id IS NOT NULL THEN
                  INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
                  VALUES 
                  (_permission_id, _role_id, now(), now(), _rs_path)
                  ON CONFLICT DO NOTHING;
              END IF;  
            END LOOP;
            END IF;
        END LOOP;
  end \$\$;
EOF
