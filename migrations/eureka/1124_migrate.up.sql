-- ALTER COLUMN exam_lo_submission.total_point
ALTER TABLE public.exam_lo_submission
  DROP COLUMN IF EXISTS total_point;

ALTER TABLE public.exam_lo_submission
  ADD COLUMN IF NOT EXISTS total_point INTEGER DEFAULT 0;

-- ALTER COLUMN exam_lo_submission_answer.point
ALTER TABLE public.exam_lo_submission_answer
  DROP COLUMN IF EXISTS point;

ALTER TABLE public.exam_lo_submission_answer
  ADD COLUMN IF NOT EXISTS point INTEGER DEFAULT 0;

-- ALTER COLUMN exam_lo_submission_score.point
ALTER TABLE public.exam_lo_submission_score
  DROP COLUMN IF EXISTS point;

ALTER TABLE public.exam_lo_submission_score
  ADD COLUMN IF NOT EXISTS point INTEGER DEFAULT 0;

-- Create trigger function for shuffled_quiz_sets
DROP TRIGGER IF EXISTS migrate_to_exam_lo_submission_fn on public.shuffled_quiz_sets;

TRUNCATE public.exam_lo_submission CASCADE;

CREATE OR REPLACE FUNCTION migrate_to_exam_lo_submission_fn()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $FUNCTION$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.shuffled_quiz_sets SQ
         WHERE SQ.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
           AND SQ.original_shuffle_quiz_set_id IS NULL
           AND EXISTS (SELECT 1
                         FROM public.study_plan_items SP
                        WHERE SP.study_plan_item_id = SQ.study_plan_item_id
                          AND SP.completed_at IS NOT NULL
                          AND EXISTS (SELECT 1 FROM public.exam_lo WHERE learning_material_id = SP.content_structure->>'lo_id'))
    )
    THEN
        INSERT INTO public.exam_lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            status,
            result,
            created_at,
            updated_at,
            deleted_at,
            total_point,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            (SELECT SP.master_study_plan_id
               FROM study_plans SP
                   LEFT JOIN study_plan_items SPI
                       ON SPI.study_plan_id = SP.study_plan_id
              WHERE SPI.study_plan_item_id = NEW.study_plan_item_id),
            (SELECT content_structure->>'lo_id' FROM public.study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id),
            NEW.shuffled_quiz_set_id,
            'SUBMISSION_STATUS_RETURNED',
            'EXAM_LO_SUBMISSION_COMPLETED',
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            COALESCE((SELECT SUM(point) FROM public.quizzes WHERE external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT exam_lo_submission_pk DO UPDATE SET
            student_id = NEW.student_id,
            study_plan_id = (SELECT SP.master_study_plan_id
                               FROM study_plans SP
                                   LEFT JOIN study_plan_items SPI
                                       ON SPI.study_plan_id = SP.study_plan_id
                              WHERE SPI.study_plan_item_id = NEW.study_plan_item_id),
            learning_material_id = (SELECT content_structure->>'lo_id' FROM public.study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id),
            shuffled_quiz_set_id = NEW.shuffled_quiz_set_id,
            status = 'SUBMISSION_STATUS_RETURNED',
            result = 'EXAM_LO_SUBMISSION_COMPLETED',
            created_at = NEW.created_at,
            updated_at = NEW.updated_at,
            deleted_at = NEW.deleted_at,
            total_point = COALESCE((SELECT SUM(point) FROM public.quizzes WHERE external_id = ANY(NEW.quiz_external_ids)), 0);
    END IF;
RETURN NULL;
END;
$FUNCTION$;

CREATE TRIGGER migrate_to_exam_lo_submission_fn
AFTER INSERT OR UPDATE ON public.shuffled_quiz_sets
FOR EACH ROW
EXECUTE FUNCTION public.migrate_to_exam_lo_submission_fn();

-- Migrate old data
INSERT INTO exam_lo_submission (
    submission_id,
    student_id,
    study_plan_id,
    learning_material_id,
    shuffled_quiz_set_id,
    status,
    result,
    created_at,
    updated_at,
    deleted_at,
    total_point,
    resource_path
)
SELECT generate_ulid() AS submission_id,
       SQ.student_id AS student_id,
       (SELECT SP.master_study_plan_id
          FROM study_plans SP
              LEFT JOIN study_plan_items SPI
                  ON SPI.study_plan_id = SP.study_plan_id
         WHERE SPI.study_plan_item_id = SQ.study_plan_item_id) AS study_plan_id,
       (SELECT content_structure->>'lo_id' FROM public.study_plan_items WHERE study_plan_item_id = SQ.study_plan_item_id) AS learning_material_id,
       SQ.shuffled_quiz_set_id AS shuffled_quiz_set_id,
       'SUBMISSION_STATUS_RETURNED' AS status,
       'EXAM_LO_SUBMISSION_COMPLETED' AS result,
       SQ.created_at AS created_at,
       SQ.updated_at AS updated_at,
       SQ.deleted_at AS deleted_at,
       COALESCE((SELECT SUM(point) FROM public.quizzes WHERE external_id = ANY(SQ.quiz_external_ids)), 0) AS total_point,
       SQ.resource_path AS resource_path
  FROM public.shuffled_quiz_sets SQ
 WHERE SQ.original_shuffle_quiz_set_id IS NULL
   AND EXISTS (SELECT 1
                FROM public.study_plan_items SP
               WHERE SP.study_plan_item_id = SQ.study_plan_item_id
                 AND SP.completed_at IS NOT NULL
                 AND EXISTS (SELECT 1 FROM public.exam_lo WHERE learning_material_id = SP.content_structure->>'lo_id'))
ON CONFLICT ON CONSTRAINT exam_lo_submission_pk DO NOTHING;
