#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update locations set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update course_access_paths set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update user_access_paths set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update granted_role_access_path set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update lesson_student_subscription_access_path set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update info_notifications_access_paths set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update lessons set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where center_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update class set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update student_enrollment_status_history set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update notification_class_members set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update notification_student_courses set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

update notification_location_filter set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

EOF

