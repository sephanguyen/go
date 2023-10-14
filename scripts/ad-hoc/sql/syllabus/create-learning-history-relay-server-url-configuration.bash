#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

RELAY_SERVER_URL=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
DO
\$\$
  BEGIN
  IF NOT EXISTS (SELECT 1 FROM public.configuration_key WHERE config_key = 'syllabus.study_plan.learning_history.relay_server_url')
  THEN
    INSERT INTO configuration_key (config_key, value_type, default_value, configuration_type, created_at, updated_at)
      VALUES ('syllabus.study_plan.learning_history.relay_server_url', 'string', '', 'CONFIGURATION_TYPE_INTERNAL', now(), now());
  END IF;
  
  UPDATE internal_configuration_value
    SET config_value = '${RELAY_SERVER_URL}'
  WHERE config_key = 'syllabus.study_plan.learning_history.relay_server_url' and resource_path IN ('-2147483630', '-2147483629'); -- Withus & WithusHS
  END;
\$\$
EOF
