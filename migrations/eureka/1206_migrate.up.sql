ALTER TABLE ONLY public.flash_card_submission
    DROP CONSTRAINT IF EXISTS flash_card_submission_shuffled_quiz_set_id_un,
    ADD CONSTRAINT flash_card_submission_shuffled_quiz_set_id_un UNIQUE (shuffled_quiz_set_id),
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS result;


CREATE OR REPLACE FUNCTION public.migrate_to_flash_card_submission_fn()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
BEGIN
    -- It's from flash card type
    -- and student actually submitted their flash card answer
    IF EXISTS(
        SELECT 1 FROM flash_card WHERE learning_material_id = NEW.learning_material_id
    )
    THEN
        INSERT INTO public.flash_card_submission (
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
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT flash_card_submission_shuffled_quiz_set_id_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point;
    END IF;
RETURN NULL;
END;   
$$;

DROP TRIGGER IF EXISTS migrate_to_flash_card_submission ON public.shuffled_quiz_sets;
CREATE TRIGGER migrate_to_flash_card_submission
AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets
FOR EACH ROW
EXECUTE FUNCTION public.migrate_to_flash_card_submission_fn();


-- Migrate data into flash_card_submission
INSERT INTO public.flash_card_submission (
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
       sqs.student_id,
       sqs.study_plan_id,
       sqs.learning_material_id,
       sqs.shuffled_quiz_set_id,
       sqs.created_at,
       sqs.updated_at,
       sqs.deleted_at,
       COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(sqs.quiz_external_ids)), 0) AS total_point,
       sqs.resource_path
FROM public.shuffled_quiz_sets sqs
WHERE sqs.deleted_at IS NULL AND sqs.updated_at > sqs.created_at
    AND EXISTS (SELECT 1 FROM public.flash_card fc WHERE fc.deleted_at IS NULL AND fc.learning_material_id = sqs.learning_material_id)
ON CONFLICT ON CONSTRAINT flash_card_submission_pk DO NOTHING;
