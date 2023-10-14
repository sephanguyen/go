#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

INSERT INTO public.class_member
(class_member_id, class_id, user_id, created_at, updated_at, deleted_at, resource_path, start_date, end_date)
VALUES('01GHT77G0QB1S8F6KC35XC1333', '01H57SN43RVA06QV7GMPG4J6X9', '01H53X9P54VZTN4QACZFFXJP5Q',
 '2022-11-14 11:44:13.486', '2022-11-14 11:44:13.486', NULL, '-2147483623', '2022-02-28 22:00:00.000', '2024-02-29 21:59:59.000')
 ON CONFLICT DO NOTHING;

INSERT INTO public.class_member
(class_member_id, class_id, user_id, created_at, updated_at, deleted_at, resource_path, start_date, end_date)
VALUES('01GHT77G0QB1S8F6KC35XC1444', '01H57SN43RVA06QV7GMPG4J6X9', '01H53XB18X3E22S6P228R7RR18',
 '2022-11-14 11:44:13.486', '2022-11-14 11:44:13.486', NULL, '-2147483623', '2022-02-28 22:00:00.000', '2024-02-29 21:59:59.000')
 ON CONFLICT DO NOTHING;

 INSERT INTO public.subject
(subject_id, "name", created_at, updated_at, deleted_at, resource_path)
VALUES('01GHT77G0QB1S8F6KC35XC1555', 'subject for test', '2022-02-28 22:00:00.000', '2022-02-28 22:00:00.000', NULL, '-2147483623')
ON CONFLICT DO NOTHING;



EOF

