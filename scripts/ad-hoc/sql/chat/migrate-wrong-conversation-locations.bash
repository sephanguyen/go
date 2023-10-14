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
    CREATE TEMP TABLE temp_wrong_deleted_conv_locs AS (
      SELECT ROW_NUMBER() OVER (ORDER BY cl.conversation_id) row_number,
          cl.conversation_id conversation_id,
          cl.location_id location_id
      FROM conversation_locations cl 
          JOIN conversation_members cm ON cl.conversation_id = cm.conversation_id
          JOIN users u ON cm.user_id = u.user_id 
          JOIN locations l ON cl.location_id = l.location_id 
          JOIN location_types lt ON lt.location_type_id = l.location_type 
      WHERE u.user_group = 'USER_GROUP_STUDENT'
          AND cl.location_id = ANY (
            SELECT uap.location_id 
            FROM user_access_paths uap 
            WHERE uap.user_id = cm.user_id 
              AND uap.deleted_at IS NULL
          )
          AND cl.deleted_at IS NOT NULL
          AND u.deleted_at IS NULL 
          AND l.deleted_at IS NULL 
          AND lt.deleted_at IS NULL 
          AND lt."name" != 'org'
    );
    SELECT INTO total_records COUNT(*) FROM temp_wrong_deleted_conv_locs;

	  WHILE counter <= total_records LOOP
        UPDATE public.conversation_locations
        SET deleted_at = NULL
        WHERE (conversation_id, location_id) IN (
            SELECT conversation_id, location_id
            FROM temp_wrong_deleted_conv_locs tmp
            WHERE tmp.row_number > counter AND tmp.row_number <= counter + batch_size
        );
        
        GET DIAGNOSTICS affected_rows = ROW_COUNT;
        COMMIT;
       
        counter := counter+batch_size;
		    total_affected_rows := total_affected_rows + affected_rows;
       	RAISE INFO '% conversation(s) processed', total_affected_rows;
	  END LOOP;
    DROP TABLE IF EXISTS temp_wrong_deleted_conv_locs;
END \$\$;
EOF
