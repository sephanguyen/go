CREATE OR REPLACE FUNCTION public.get_student_completion_learning_material()
    RETURNS table
        (
            study_plan_id        text,
            student_id           text,
            learning_material_id text,
            completed_at         timestamptz
        )
    LANGUAGE sql
    stable
AS
$$
(
    SELECT sel.study_plan_id,
        sel.student_id,
        sel.learning_material_id,
        max(sel.created_at) as completed_at
    FROM student_event_logs sel
    JOIN learning_material lm USING(learning_material_id)
    LEFT JOIN exam_lo_submission els USING(study_plan_id, student_id, learning_material_id)
    WHERE 
        sel.payload ->> 'event' = ANY(ARRAY['completed', 'exited'])
        AND CASE
                -- Exam LO type:
                WHEN lm.type = 'LEARNING_MATERIAL_EXAM_LO' THEN 
                    els.submission_id IS NOT NULL AND els.deleted_at IS NULL
                -- LO, FLASH_CARD type
                ELSE TRUE
            END
    GROUP BY sel.study_plan_id, sel.student_id, sel.learning_material_id
)
UNION
(
    SELECT study_plan_id,
        student_id,
        learning_material_id,
        max(complete_date) as completed_at
    FROM student_submissions
    WHERE deleted_at IS NULL
    GROUP BY study_plan_id, student_id, learning_material_id
)
$$;
