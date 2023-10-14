#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

------------ KEC LMS Pilot -------------
INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH751')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH752')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH753')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH754')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH755')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH756')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;
INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH757')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH758')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH759')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD6YPRS45EZ23482PFH750')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD78816KTXZJ2WC9Q2VGP1')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GTBD78816KTXZJ2WC9Q2VGP2')
ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;
EOF
