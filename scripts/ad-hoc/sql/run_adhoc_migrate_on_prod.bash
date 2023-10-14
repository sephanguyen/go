#!/bin/bash

# This script run file migrate `migrations/lessonmgmt/1087_migrate.up.sql` drop fk of destination table

set -euo pipefail

DB_NAME="lessonmgmt"


psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
    ALTER TABLE public.lesson_members_states DROP CONSTRAINT IF EXISTS lesson_id_fk;
    ALTER TABLE public.lesson_room_states DROP CONSTRAINT IF EXISTS lesson_id_fk;
EOF
