#!/bin/bash

# This script migrate Timesheet service setting value for Partner in configturation table.

set -euo pipefail

DB_MASTERMGMT_NAME="mastermgmt"
DB_TIMESHEET_NAME="timesheet"

resource_path=""

print_usage() {
  printf "Usage: Update data auto create flag, auto create flag log"
}

while getopts 'r:' flag; do
  case "${flag}" in
    r) resource_path="${OPTARG}" ;;
    *) print_usage
       exit 1 ;;
  esac
done

VAR=$(psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_MASTERMGMT_NAME}" -p "${DB_PORT}" -c "select       
  (case when exists (SELECT * FROM internal_configuration_value WHERE config_value = 'off' and config_key = 'hcm.timesheet_management'
    and deleted_at IS NULL
    and resource_path = '$resource_path')
    then true 
    else false
  end) as rs;")

#result when query is true: rs -------- t (1 row)
#result when query is false: rs -------- f (1 row)

if [[ $VAR == *"t"* ]]; then
  
  psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_TIMESHEET_NAME}" -p "${DB_PORT}" \ -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
  BEGIN;

    --update value of auto create flag to off
    UPDATE auto_create_timesheet_flag SET flag_on = false, updated_at=NOW() WHERE deleted_at IS NULL and resource_path = '$resource_path';

    --add log when change auto create flag value
    DO \$\$
      DECLARE
      total_staff int;
      counter int := 0;

      BEGIN

        SELECT INTO total_staff COUNT(DISTINCT staff_id) FROM auto_create_timesheet_flag WHERE deleted_at IS NULL and resource_path = '$resource_path';
        
        WHILE counter <= total_staff LOOP
          INSERT INTO auto_create_flag_activity_log (id, staff_id, change_time, flag_on, created_at, updated_at, resource_path)
            SELECT gen_random_uuid(), ac.staff_id, NOW(), false, NOW(), NOW(), '$resource_path'
            FROM auto_create_timesheet_flag ac WHERE deleted_at IS NULL and resource_path = '$resource_path'
            ORDER BY ac.created_at asc LIMIT 1 OFFSET counter;            
          
          counter := counter+1;

        END LOOP;
    END \$\$;

  COMMIT;

EOF
fi
