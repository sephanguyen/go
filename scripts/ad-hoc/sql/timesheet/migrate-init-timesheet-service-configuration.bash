#!/bin/bash

# This script init Timesheet service setting value for Partner in configturation table.
# Example: ./scripts/ad-hoc/sql/timesheet/migrate-init-timesheet-service-configuration.bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
    INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
    values
    ('01GHB11GT1RY9VYT61AFB1J1FK', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483648'), --Manabie
    ('01GHB11HZ95FEW10FN469C6519', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483647'), --JPREP
    ('01GHB11HZ95FEW10FN48N2HSJG', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483646'), --Synersia
    ('01GHB11HZ96CD6PKCZ7EZG0YFW', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483645'), --Renseikai
    ('01GHB11HZ9ANBWJ9F2ZNV1WAB9', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483644'), --Bestco
    ('01GHB11HZ9MTQ4HS5R9B3E6Z7D', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483643'), --Bestco
    ('01GHB11HZA7BJWNNS3H4NEPK86', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483642'), --KEC
    ('01GHB11HZSWBBX2ERSJ9VW8FWF', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483641'), --AIC
    ('01GHB11HZY07NYMPHPPHAHW5KG', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483640'), --NSG
    ('01GHB11HZZP4RBJ4M9385GNVYX', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483636'), --withus
    ('01GHB11J03W8SEAXR1NJY2TCX3', 'hcm.timesheet_management', 'string', now(), now(), 'on', '-2147483635'), --KEC
    ('01GHB11J03W8SEAXR1NPHSTXG1', 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483631') --Eishinkan
    ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO UPDATE SET config_value = EXCLUDED.config_value
    ; 
EOF
