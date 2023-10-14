#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE info_notifications dest SET receiver_names = src.receiver_names
FROM (
	SELECT info_noti.notification_id , array_agg(u."name") as receiver_names
	FROM users u
	join (
		SELECT in2.notification_id, unnest(in2.generic_receiver_ids) as user_id
		FROM info_notifications in2
	) info_noti on info_noti.user_id = u.user_id
	group by info_noti.notification_id
) src
where dest.notification_id = src.notification_id AND dest.generic_receiver_ids IS NOT NULL;
EOF
