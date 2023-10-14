#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483648') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483647') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483646') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483645') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483644') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483643') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483642') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483641') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483640') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483639') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483638') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483637') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483636') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483635') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483634') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483633') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483632') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483631') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483630') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.is_enabled', 'boolean', now(), now(), 'false', '-2147483629') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;




-- ZOOM-CONFIG
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483648') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483647') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483646') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483645') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483644') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483643') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483642') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483641') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483640') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483639') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483638') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483637') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483636') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483635') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483634') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483633') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483632') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483631') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'json', now(), now(), '{}', '-2147483630') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) values
(uuid_generate_v4(), 'lesson.zoom.config', 'boolean', now(), now(), '{}', '-2147483629') 
ON CONFLICT ON CONSTRAINT external_configuration_value_resource_unique DO NOTHING;




EOF

