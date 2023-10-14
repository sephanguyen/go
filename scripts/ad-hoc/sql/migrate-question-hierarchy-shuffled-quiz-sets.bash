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
    CREATE TEMP TABLE temp_shuffled_quiz_sets AS (
        SELECT
            ROW_NUMBER() OVER (ORDER BY sqs.shuffled_quiz_set_id) row_number,
            sqs.shuffled_quiz_set_id,
            sqs.quiz_external_ids
        FROM shuffled_quiz_sets sqs
        INNER JOIN quiz_sets qs on sqs.original_quiz_set_id = qs.quiz_set_id
      	LEFT JOIN flash_card fc on qs.lo_id = fc.learning_material_id
        WHERE fc.learning_material_id IS NULL
        AND ARRAY_LENGTH(sqs.question_hierarchy, 1) IS NULL AND ARRAY_LENGTH(sqs.quiz_external_ids, 1) > 0
    );
    SELECT INTO total_records COUNT(*) FROM temp_shuffled_quiz_sets;

	WHILE counter <= total_records LOOP
        UPDATE public.shuffled_quiz_sets
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
        WHERE shuffled_quiz_set_id IN (
            SELECT shuffled_quiz_set_id
            FROM temp_shuffled_quiz_sets tqs
            WHERE tqs.row_number > counter AND tqs.row_number <= counter + batch_size
        );
        
        GET DIAGNOSTICS affected_rows = ROW_COUNT;
        COMMIT;
       	counter := counter+batch_size;
		total_affected_rows := total_affected_rows + affected_rows;
       	RAISE INFO '% rows processed', total_affected_rows;
	END LOOP;
    DROP TABLE IF EXISTS temp_shuffled_quiz_sets;
END \$\$;
EOF
