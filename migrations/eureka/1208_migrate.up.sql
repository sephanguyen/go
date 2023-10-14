DROP TRIGGER IF EXISTS migrate_to_lo_submission ON public.shuffled_quiz_sets;
TRUNCATE lo_submission CASCADE;

CREATE OR REPLACE FUNCTION migrate_to_lo_submission_fn()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM public.learning_objective LO 
        WHERE LO.learning_material_id = NEW.learning_material_id
    )
    THEN
        INSERT INTO public.lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            total_point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            COALESCE((SELECT SUM(point) FROM public.quizzes WHERE quizzes.deleted_at IS NULL AND quizzes.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point;
    END IF;
RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS migrate_to_lo_submission ON public.shuffled_quiz_sets;
CREATE TRIGGER migrate_to_lo_submission
AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets
FOR EACH ROW
EXECUTE FUNCTION public.migrate_to_lo_submission_fn();

-- Migrate old data
INSERT INTO lo_submission (
    submission_id,
    student_id,
    study_plan_id,
    learning_material_id,
    shuffled_quiz_set_id,
    created_at,
    updated_at,
    deleted_at,
    total_point,
    resource_path
)
SELECT generate_ulid() AS submission_id,
    SQ.student_id,
    SQ.study_plan_id,
    SQ.learning_material_id,
    SQ.shuffled_quiz_set_id,
    SQ.created_at,
    SQ.updated_at,
    SQ.deleted_at,
    COALESCE((SELECT SUM(point) FROM public.quizzes WHERE quizzes.deleted_at IS NULL AND quizzes.external_id = ANY(SQ.quiz_external_ids)), 0) AS total_point,
    SQ.resource_path
FROM public.shuffled_quiz_sets SQ
INNER JOIN public.learning_objective LO USING(learning_material_id)
WHERE SQ.updated_at > SQ.created_at --check student submited or not
    AND SQ.deleted_at IS NULL 
    AND LO.deleted_at IS NULL 
ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO NOTHING;
