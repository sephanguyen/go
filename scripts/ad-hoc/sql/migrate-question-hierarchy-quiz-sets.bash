#!/bin/bash

set -euo pipefail

DB_NAME="eureka"

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
    CREATE TEMP TABLE temp_quiz_sets AS (
        SELECT
            ROW_NUMBER() OVER (ORDER BY qs.quiz_set_id) row_number,
            qs.quiz_set_id,
            qs.quiz_external_ids
        FROM quiz_sets qs
        WHERE ARRAY_LENGTH(qs.question_hierarchy, 1) IS NULL AND ARRAY_LENGTH(qs.quiz_external_ids, 1) > 0
    );

    CREATE TEMP TABLE temp_flash_card AS (
        SELECT learning_material_id
        FROM flash_card
        WHERE deleted_at IS NULL
    );

    SELECT INTO total_records COUNT(*) FROM temp_quiz_sets;

	WHILE counter <= total_records LOOP
        UPDATE public.quiz_sets
        SET question_hierarchy = (
            CASE WHEN ARRAY_LENGTH(quiz_external_ids, 1) IS NULL
            THEN ARRAY[]::JSONB[]
            ELSE
            (
                SELECT ARRAY_AGG(
                    TO_JSONB(qei)
                ) FROM (
                    SELECT UNNEST(quiz_external_ids) as id, 'QUESTION' as type
                ) qei
            )
            END
        )
        WHERE quiz_set_id IN (
            SELECT quiz_set_id
            FROM temp_quiz_sets tqs
            WHERE tqs.row_number > counter AND tqs.row_number <= counter + batch_size
        )
        AND NOT EXISTS(SELECT 1 FROM temp_flash_card tfc WHERE tfc.learning_material_id=lo_id);

        GET DIAGNOSTICS affected_rows = ROW_COUNT;
        COMMIT;
       	counter := counter+batch_size;
		total_affected_rows := total_affected_rows + affected_rows;
       	RAISE INFO '% rows processed', total_affected_rows;
	END LOOP;
    DROP TABLE IF EXISTS temp_quiz_sets;
    DROP TABLE IF EXISTS temp_flash_card;
END \$\$;
EOF
