#!/bin/bash

set -euo pipefail

DB_TARGET_NAME="mastermgmt"


psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_TARGET_NAME}" -p "${DB_PORT}"  --command="

BEGIN;

UPDATE public.location_configuration_value
set config_value=true
WHERE config_key=any('{communication.chat.enable_student_chat, communication.chat.enable_parent_chat}') and resource_path = '-2147483646';

COMMIT;"
