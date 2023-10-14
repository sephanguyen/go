-- Drop the deprecated triggers
DROP TRIGGER IF EXISTS migrate_to_exam_lo_submission_fn on public.shuffled_quiz_sets;
DROP TRIGGER IF EXISTS trigger_study_plan_items_completed_at_to_exam_lo_submission_fn on public.study_plan_items;

DROP TRIGGER IF EXISTS migrate_to_exam_lo_submission_and_answer_once_submitted on public.shuffled_quiz_sets;

TRUNCATE exam_lo_submission CASCADE;

CREATE OR REPLACE FUNCTION public.migrate_to_exam_lo_submission_and_answer_once_submitted()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $FUNCTION$
DECLARE
    _submission_id text;
BEGIN
    -- It's from Exam LO type
    -- and student actually submitted their answer (trigger after update of updated_at)
    IF EXISTS(
        SELECT 1 FROM exam_lo WHERE learning_material_id = NEW.learning_material_id
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
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            'SUBMISSION_STATUS_RETURNED',
            'EXAM_LO_SUBMISSION_COMPLETED',
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_un DO UPDATE SET
            student_id = EXCLUDED.student_id,
            study_plan_id = EXCLUDED.study_plan_id,
            learning_material_id = EXCLUDED.learning_material_id,
            status = EXCLUDED.status,
            result = EXCLUDED.result,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point
        RETURNING submission_id into _submission_id;

        INSERT INTO public.exam_lo_submission_answer (
            student_id,
            submission_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            quiz_id,
            student_text_answer,
            correct_text_answer,
            student_index_answer,
            correct_index_answer,
            is_correct,
            is_accepted,
            point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
        SELECT NEW.student_id,
               _submission_id AS submission_id,
               NEW.study_plan_id,
               NEW.learning_material_id,
               NEW.shuffled_quiz_set_id,
               SA.quiz_id,
               (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'filled_text')::jsonb) X(obj)) AS student_text_answer,
               (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_text')::jsonb) X(obj)) AS correct_text_answer,
               (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'selected_index')::jsonb) X(obj))::INTEGER[] AS student_index_answer,
               (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_index')::jsonb) X(obj))::INTEGER[] AS correct_index_answer,
               ARRAY(SELECT jsonb_array_elements_text((SA.quiz_history->>'correctness')::jsonb))::BOOLEAN[] AS is_correct,
               (SA.quiz_history->>'is_accepted')::BOOLEAN AS is_accepted,
               CASE WHEN SA.quiz_history IS NOT NULL THEN
                   COALESCE((SA.quiz_history->>'is_accepted')::BOOLEAN::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = SA.quiz_id), 0)
                ELSE 0
               END AS point, -- If there is no answer for the question, then 0 point as default.
               NEW.created_at,
               NEW.updated_at,
               NEW.deleted_at,
               NEW.resource_path
          -- The table contains the latest quiz_history by quiz_id, which in submission_history column.
          FROM (SELECT quiz_id,
                       (SELECT DISTINCT ON (X.obj ->> 'quiz_id') X.obj
                          FROM public.shuffled_quiz_sets Y
                              CROSS JOIN jsonb_array_elements(Y.submission_history) X(obj)
                         WHERE Y.shuffled_quiz_set_id = SQ.shuffled_quiz_set_id
                           AND X.obj ->> 'quiz_id' = SQ.quiz_id
                         ORDER BY X.obj ->> 'quiz_id', X.obj->>'submitted_at' DESC) AS quiz_history
                  -- For each record in exam_lo_submission table, there will be respectively n records in exam_lo_submission_answer table
                  -- based on shuffled_quiz_sets.quiz_external_ids column.
                  -- In case of quiz_external_ids is null, there is no record created (UNNEST deals with it).
                  FROM (SELECT shuffled_quiz_set_id,
                               UNNEST(quiz_external_ids) AS quiz_id
                          FROM public.shuffled_quiz_sets
                         WHERE shuffled_quiz_set_id = NEW.shuffled_quiz_set_id) SQ
               ) SA
        ON CONFLICT ON CONSTRAINT exam_lo_submission_answer_pk DO UPDATE SET
            study_plan_id = EXCLUDED.study_plan_id,
            learning_material_id = EXCLUDED.learning_material_id,
            shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id,
            student_text_answer = EXCLUDED.student_text_answer,
            correct_text_answer = EXCLUDED.correct_text_answer,
            student_index_answer = EXCLUDED.student_index_answer,
            correct_index_answer = EXCLUDED.correct_index_answer,
            is_correct = EXCLUDED.is_correct,
            is_accepted = EXCLUDED.is_accepted,
            point = EXCLUDED.point,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at;
    END IF;
RETURN NULL;
END;
$FUNCTION$;

DROP TRIGGER IF EXISTS migrate_to_exam_lo_submission_and_answer_once_submitted ON public.shuffled_quiz_sets;
CREATE TRIGGER migrate_to_exam_lo_submission_and_answer_once_submitted
AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets
FOR EACH ROW
EXECUTE FUNCTION public.migrate_to_exam_lo_submission_and_answer_once_submitted();

-- Migrate data into exam_lo_submission
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
SELECT generate_ulid() AS submission_id,
       SQ.student_id,
       SQ.study_plan_id,
       SQ.learning_material_id,
       SQ.shuffled_quiz_set_id,
       'SUBMISSION_STATUS_RETURNED' AS status,
       'EXAM_LO_SUBMISSION_COMPLETED' AS result,
       SQ.created_at,
       SQ.updated_at,
       SQ.deleted_at,
       COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(SQ.quiz_external_ids)), 0) AS total_point,
       SQ.resource_path
  FROM public.shuffled_quiz_sets SQ
 WHERE SQ.updated_at > SQ.created_at -- Once student actually submitted their answer for Exam LO.
   AND EXISTS (SELECT 1 FROM public.exam_lo EL WHERE EL.deleted_at IS NULL AND EL.learning_material_id = SQ.learning_material_id)
ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_un DO NOTHING;

-- Migrate data into exam_lo_submission_answer
INSERT INTO public.exam_lo_submission_answer (
    student_id,
    submission_id,
    study_plan_id,
    learning_material_id,
    shuffled_quiz_set_id,
    quiz_id,
    student_text_answer,
    correct_text_answer,
    student_index_answer,
    correct_index_answer,
    is_correct,
    is_accepted,
    point,
    created_at,
    updated_at,
    deleted_at,
    resource_path
)
SELECT SA.student_id,
       SA.submission_id,
       SA.study_plan_id,
       SA.learning_material_id,
       SA.shuffled_quiz_set_id,
       SA.quiz_id,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'filled_text')::jsonb) X(obj)) AS student_text_answer,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_text')::jsonb) X(obj)) AS correct_text_answer,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'selected_index')::jsonb) X(obj))::INTEGER[] AS student_index_answer,
       (SELECT array_agg(obj) FROM jsonb_array_elements_text((SA.quiz_history->>'correct_index')::jsonb) X(obj))::INTEGER[] AS correct_index_answer,
       ARRAY(SELECT jsonb_array_elements_text((SA.quiz_history->>'correctness')::jsonb))::BOOLEAN[] AS is_correct,
       (SA.quiz_history->>'is_accepted')::BOOLEAN AS is_accepted,
       CASE WHEN SA.quiz_history IS NOT NULL THEN
           COALESCE((SA.quiz_history->>'is_accepted')::BOOLEAN::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = SA.quiz_id), 0)
        ELSE 0
       END AS point, -- If there is no answer for the question, then 0 point as default.
       created_at,
       updated_at,
       deleted_at,
       resource_path
  FROM (SELECT SE.student_id,
               SE.submission_id,
               SE.study_plan_id,
               SE.learning_material_id,
               SE.shuffled_quiz_set_id,
               SE.quiz_id,
               (SELECT DISTINCT ON (X.obj ->> 'quiz_id') X.obj
                  FROM public.shuffled_quiz_sets Y
                      CROSS JOIN jsonb_array_elements(Y.submission_history) X(obj)
                 WHERE Y.shuffled_quiz_set_id = SE.shuffled_quiz_set_id
                   AND X.obj ->> 'quiz_id' = SE.quiz_id
                 ORDER BY X.obj ->> 'quiz_id', X.obj->>'submitted_at' DESC) AS quiz_history,
               SE.created_at,
               SE.updated_at,
               SE.deleted_at,
               SE.resource_path
          -- For each record on exam_lo_submission table, there will be respectively n records on exam_lo_submission_answer table
          -- based on shuffled_quiz_sets.quiz_external_ids column.
          -- In case of quiz_external_ids is null, there is no record created (UNNEST deals with it).
          FROM (SELECT SQ.student_id,
                       EL.submission_id,
                       SQ.study_plan_id,
                       SQ.learning_material_id,
                       SQ.shuffled_quiz_set_id,
                       UNNEST(quiz_external_ids) AS quiz_id,
                       SQ.created_at,
                       SQ.updated_at,
                       SQ.deleted_at,
                       SQ.resource_path
                  FROM public.shuffled_quiz_sets SQ
                    JOIN exam_lo_submission EL ON EL.shuffled_quiz_set_id = SQ.shuffled_quiz_set_id -- Which are existed on exam_lo_submission table.
               ) SE
       ) SA
ON CONFLICT ON CONSTRAINT exam_lo_submission_answer_pk DO NOTHING;