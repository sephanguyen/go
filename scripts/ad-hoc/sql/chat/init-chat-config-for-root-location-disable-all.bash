#!/bin/bash

set -euo pipefail

DB_SOURCE_NAME="bob"
DB_TARGET_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_SOURCE_NAME}" -p "${DB_PORT}" \
--command="copy (select l.location_id, l.resource_path from locations l
join location_types lt on l.location_type = lt.location_type_id
where lt.name = 'org') to stdout with csv" | \
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_TARGET_NAME}" -p "${DB_PORT}"  --command="

-- hold root location ids from bob to TEMPORARY TABLE
CREATE TEMPORARY TABLE org_ids(
   location_id text,
   resource_path text
);
copy org_ids (location_id, resource_path) from stdin csv;

BEGIN;
-- insert 4 conifig keys, this action will aslo create records in external_configuration_value table
INSERT INTO configuration_key (config_key, value_type, configuration_type, created_at, updated_at) values
('communication.chat.enable_student_chat', 'boolean' ,'CONFIGURATION_TYPE_EXTERNAL', now(), now()),
('communication.chat.enable_parent_chat', 'boolean' ,'CONFIGURATION_TYPE_EXTERNAL', now(), now());

-- insert default configs in org_ids to location_configuration_value table
-- insert key 'communication.chat.enable_student_chat'
INSERT INTO location_configuration_value (location_config_id, config_key, location_id, config_value_type, config_value, created_at, updated_at,  resource_path)
SELECT uuid_generate_v4() AS uuid_generate_v4, e.config_key, o.location_id, e.config_value_type, 'false', now(), now(), o.resource_path from org_ids o join
external_configuration_value e on o.resource_path = e.resource_path
where e.config_key = 'communication.chat.enable_student_chat';

-- insert key 'communication.chat.enable_parent_chat'
INSERT INTO location_configuration_value (location_config_id, config_key, location_id, config_value_type, config_value, created_at, updated_at,  resource_path)
SELECT uuid_generate_v4() AS uuid_generate_v4, e.config_key, o.location_id, e.config_value_type, 'false', now(), now(), o.resource_path from org_ids o join
external_configuration_value e on o.resource_path = e.resource_path
where e.config_key = 'communication.chat.enable_parent_chat';

COMMIT;"
