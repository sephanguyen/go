#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE info_notifications in2 SET generic_receiver_ids = res.expected_generic_ids
FROM (
	WITH cte2 AS (
		WITH cte AS (
			SELECT in3.notification_id ,
				t.user_group,
				in3.receiver_ids ,
				(
					CASE 
						WHEN t.user_group = 'USER_GROUP_PARENT'
						THEN (
							SELECT array_agg(DISTINCT sp.parent_id)
							FROM student_parents sp
							WHERE sp.student_id = t2.receiver_id
						)
						WHEN t.user_group = 'USER_GROUP_STUDENT'
						THEN (
							SELECT array_agg(DISTINCT u.user_id) 
							FROM users u
							WHERE u.user_id = t2.receiver_id
							AND u.user_group = 'USER_GROUP_STUDENT'
						)
					END
				) AS expected
			FROM info_notifications in3 
			-- table t stores notification and it's user_group
			JOIN (
				SELECT in2.notification_id ,
					jsonb_array_elements_text(in2.target_groups -> 'user_group_filter'->'user_group') AS user_group
				FROM info_notifications in2
				WHERE (in2.receiver_ids IS NOT NULL AND in2.receiver_ids <> '{}')
					AND in2.target_groups -> 'user_group_filter'->>'user_group' IS NOT null
					AND jsonb_array_length(in2.target_groups -> 'user_group_filter'->'user_group') > 0
			) t ON t.notification_id = in3.notification_id
			-- table t2 stores notification and it's receiver_ids
			JOIN (
				SELECT in4.notification_id,
					unnest(in4.receiver_ids) AS receiver_id
				FROM info_notifications in4 
				WHERE (in4.receiver_ids IS NOT NULL AND in4.receiver_ids <> '{}')
					AND in4.target_groups -> 'user_group_filter'->>'user_group' IS NOT null
					AND jsonb_array_length(in4.target_groups -> 'user_group_filter'->'user_group') > 0
			) t2 ON t2.notification_id = in3.notification_id
		)
		SELECT cte.notification_id,
			cte.receiver_ids,
			cte.user_group,
			unnest(cte.expected) AS expected_generic_ids
		FROM cte
	)
	SELECT cte2.notification_id,
		array_agg(DISTINCT cte2.expected_generic_ids) AS expected_generic_ids
	FROM cte2
	JOIN info_notifications in5 ON in5.notification_id = cte2.notification_id
	GROUP BY cte2.notification_id, cte2.receiver_ids, in5.generic_receiver_ids
) res
WHERE res.notification_id = in2.notification_id;
EOF
