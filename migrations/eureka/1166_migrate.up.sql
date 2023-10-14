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
    SELECT study_plan_id,
        student_id,
        learning_material_id,
        max(created_at) as completed_at
    FROM student_event_logs
    WHERE payload ->> 'event' = ANY(ARRAY['completed', 'exited'])
    GROUP BY study_plan_id, student_id, learning_material_id
)
UNION ALL
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
