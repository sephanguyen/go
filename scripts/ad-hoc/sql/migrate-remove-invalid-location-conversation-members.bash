#!/bin/bash

set -euo pipefail

DB_NAME="tom"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --echo-all <<EOF
DO \$\$
DECLARE
total_records int;
batch_size int := 5000;
counter int := 0;
affected_rows int;
total_affected_rows int := 0;
BEGIN
    CREATE TEMP TABLE temp_removed_users AS (
        SELECT ROW_NUMBER() OVER (ORDER BY cm.user_id) row_number,
        cm.user_id user_id,
        cm.conversation_id conversation_id
        FROM conversation_members cm
        INNER JOIN conversations cs ON cm.conversation_id = cs.conversation_id
        WHERE (cm.user_id, cm.conversation_id) NOT IN (
            SELECT DISTINCT cm2.user_id, cm2.conversation_id
            FROM conversation_members cm2
            INNER JOIN conversations cs2 ON cm2.conversation_id = cs2.conversation_id
            INNER JOIN granted_permissions gp ON gp.user_id = cm2.user_id
            INNER JOIN conversation_locations cl ON cm2.conversation_id = cl.conversation_id AND cl.location_id = gp.location_id
            WHERE cm2.status = 'CONVERSATION_STATUS_ACTIVE'
                AND cm2.role = 'USER_GROUP_TEACHER'
                AND cs2.conversation_type IN ('CONVERSATION_STUDENT', 'CONVERSATION_PARENT')
                AND cl.deleted_at IS NULL 
                AND gp.permission_name = 'master.location.read'
        )
            AND cm.status = 'CONVERSATION_STATUS_ACTIVE'
            AND cm.role = 'USER_GROUP_TEACHER'
            AND cs.conversation_type IN ('CONVERSATION_STUDENT', 'CONVERSATION_PARENT')
    );
    SELECT INTO total_records COUNT(*) FROM temp_removed_users;

	WHILE counter <= total_records LOOP
        UPDATE public.conversation_members
        SET status = 'CONVERSATION_STATUS_INACTIVE'
        WHERE (user_id, conversation_id) IN (
            SELECT user_id, conversation_id
            FROM temp_removed_users tru
            WHERE tru.row_number > counter AND tru.row_number <= counter + batch_size
        );
        
        GET DIAGNOSTICS affected_rows = ROW_COUNT;
        COMMIT;
       	counter := counter+batch_size;
		total_affected_rows := total_affected_rows + affected_rows;
       	RAISE INFO '% rows processed', total_affected_rows;
	END LOOP;
    DROP TABLE IF EXISTS temp_removed_users;
END \$\$;
EOF
