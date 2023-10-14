#!/bin/bash

# This script use to update cut off date in Timesheet confirmation window for Partner

resource_path=""
endDateOfLatestPeriodConvert=""

print_usage() {
  printf "Usage: Update cut-off date for Partner
    example: ./scripts/ad-hoc/sql/timesheet/migrate-update-cut-off-date-for-partner.bash -r -2147483648 -d 10 -n '2023-02-12 15:00:00'

    -r: resource_path or Partner
    -d: cut-off date of Partner (Ask PM or Partner for that)
    -n: in start of next cut off date effect:
      ex: cut off date is 12 and now is 2023-02-15 => n = 2023-03-12 15:00:00 +00:00 (UTC timezone)
  "
}

while getopts 'r:d:n:' flag; do
  case "${flag}" in
    r) resource_path="${OPTARG}" ;;
    d) new_cut_off_date="${OPTARG}" ;;
    n) start_next_cut_off_date="${OPTARG}" ;;
    *) print_usage
       exit 1 ;;
  esac
done


  # get latest cut-off-date ID value of partner
latestCutOffDateID=$(psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" -AXqtc "select       
  id from timesheet_confirmation_cut_off_date where deleted_at is null AND resource_path ='$resource_path' 
  ORDER BY created_at DESC LIMIT 1")

if [[ $latestCutOffDateID == "" ]]; then
  echo "cut off date for this Partner is not inited, we can not update now"
  exit
fi


  # get lastes period end date
latestPeriodEndDate=$(psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" -AXqtc "select       
  end_date from timesheet_confirmation_period where deleted_at is null AND resource_path ='$resource_path' order by end_date DESC LIMIT 1")

endDateOfLatestPeriodConvert="${start_next_cut_off_date/15:00:00/14:59:59}"    

endDateOfLatestCutOffDateConvert="${latestPeriodEndDate/14:59:59/15:00:00}"

  # update end_date of latest cut-off date row
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

    -- for run uuid_generate_v4 function
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    BEGIN;

    -- update end_date of latest cut-off date row

    UPDATE public.timesheet_confirmation_cut_off_date
    SET end_date = '$endDateOfLatestCutOffDateConvert'
    WHERE id ='$latestCutOffDateID';

    -- create new value of cut off date

    INSERT INTO public."timesheet_confirmation_cut_off_date" 
    (id, cut_off_date, start_date, end_date, created_at, updated_at, resource_path) 
    VALUES 
    (uuid_generate_v4(), '$new_cut_off_date', '$start_next_cut_off_date', '2099-12-30 14:59:59 +00:00', now(), now(), '$resource_path');

    -- insert small period from latest period with old cut off date to new value of cut off date

    INSERT INTO public."timesheet_confirmation_period" 
    (id, start_date, end_date, created_at, updated_at, resource_path) 
    VALUES 
    (uuid_generate_v4(), '$endDateOfLatestCutOffDateConvert', '$endDateOfLatestPeriodConvert', now(), now(), '$resource_path');
    
    COMMIT;

EOF

