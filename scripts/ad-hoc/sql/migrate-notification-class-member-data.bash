#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --echo-all <<EOF
\dt
DO \$\$
DECLARE
total_records int;
batch_size int := 5000;
counter int := 0;
affected_rows int := 0;
total_affected_rows int := 0;
BEGIN
	CREATE TEMP TABLE temp_class_member AS (
        SELECT ROW_NUMBER() OVER (ORDER BY cm.class_id) row_number, 
            cm.user_id AS student_id, 
            cm.class_id, 
            CASE 
                WHEN cm.start_date = '0001-01-01 07:00:00.000 +0700' THEN NULL
                ELSE cm.start_date
            END AS start_at,
            CASE 
                WHEN cm.end_date = '0001-01-01 07:00:00.000 +0700' THEN NULL
                ELSE cm.end_date
            END AS end_at,
            cm.created_at, 
            now() AS updated_at, 
            cm.resource_path, 
            nsc.location_id, 
            c.course_id
        FROM class_member cm 
            JOIN "class" c ON cm.class_id = c.class_id 
            JOIN notification_student_courses nsc ON nsc.course_id = c.course_id AND nsc.student_id = cm.user_id 
        WHERE cm.deleted_at IS NULL 
            AND c.deleted_at IS NULL 
            AND nsc.deleted_at IS NULL 
    );

    SELECT INTO total_records COUNT(*) FROM temp_class_member;

	WHILE counter <= total_records LOOP
		INSERT INTO public.notification_class_members
            (student_id, class_id, start_at, end_at, created_at, updated_at, resource_path, location_id, course_id, deleted_at)
        SELECT student_id, class_id, start_at, end_at, created_at, updated_at, resource_path, location_id, course_id, NULL
        FROM temp_class_member
        WHERE temp_class_member.row_number > counter AND temp_class_member.row_number <= counter + batch_size
        ON CONFLICT ON CONSTRAINT pk__notification_class_members DO NOTHING;
	    
	   	GET DIAGNOSTICS affected_rows = ROW_COUNT;
	    COMMIT;
	   
	   	counter := counter+batch_size;
		total_affected_rows := total_affected_rows + affected_rows;
       
        RAISE INFO '% rows inserted', total_affected_rows;
	END LOOP;
	
    DROP TABLE IF EXISTS temp_class_member;
END \$\$;
EOF
