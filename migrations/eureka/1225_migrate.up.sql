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
    SELECT
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
        COALESCE(SUM(point), 0),
        NEW.resource_path
    FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(NEW.quiz_external_ids)
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

-- upsert exam_lo_submission_answer table
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
        submitted_keys_answer,
        correct_keys_answer,
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
        sh.quiz_id,
        sh.filled_text AS student_text_answer,
        sh.correct_text AS correct_text_answer,
        sh.selected_index AS student_index_answer,
        sh.correct_index AS correct_index_answer,
        sh.submitted_keys AS submitted_keys_answer,
        sh.correct_keys AS correct_keys_answer,
        sh.correctness AS is_correct,
        sh.is_accepted AS is_accepted,
        COALESCE((sh.is_accepted)::BOOLEAN::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = sh.quiz_id LIMIT 1), 0) as point,
        NEW.created_at,
        NEW.updated_at,
        NEW.deleted_at,
        NEW.resource_path
       -- The table contains the latest quiz_history by quiz_id, which in submission_history column.
    FROM shuffled_quiz_sets AS sqs, jsonb_to_recordset(sqs.submission_history) AS sh (
        quiz_id text,
        correctness BOOLEAN[],
        filled_text TEXT[],
        is_accepted BOOLEAN,
        correct_text TEXT[],
        submitted_at timestamp with time zone,
        correct_index INTEGER[],
        selected_index INTEGER[],
        submitted_keys TEXT[],
        correct_keys TEXT[])
    WHERE sqs.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
    ON CONFLICT ON CONSTRAINT exam_lo_submission_answer_pk DO UPDATE SET
        study_plan_id = EXCLUDED.study_plan_id,
        learning_material_id = EXCLUDED.learning_material_id,
        shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id,
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer,
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
