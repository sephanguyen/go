CREATE OR REPLACE FUNCTION public.get_submission_history() 
RETURNS TABLE (
    student_id text,
    shuffled_quiz_set_id text,
    quiz_id text,
    student_TEXT_answer text[],
    correct_TEXT_answer text[],
    student_index_answer integer[],
    correct_index_answer integer[],
    is_correct BOOLEAN[],
    is_accepted BOOLEAN,
    point integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource_path text
)
LANGUAGE SQL STABLE
SECURITY INVOKER
AS $$
    SELECT DISTINCT ON (shuffled_quiz_set_id, quiz_id)
 		student_id,
        shuffled_quiz_set_id,
        sh.quiz_id,
        sh.filled_text AS student_text_answer,
		sh.correct_text AS correct_text_answer,
        sh.selected_index AS student_index_answer,
        sh.correct_index AS correct_index_answer,
       	sh.correctness AS is_correct,
        sh.is_accepted AS is_accepted,
		COALESCE((sh.is_accepted)::BOOLEAN::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = sh.quiz_id LIMIT 1), 0) as point,
        (sh.submitted_at)::timestamp with time zone as created_at,
		updated_at,
		deleted_at,
		resource_path
    FROM shuffled_quiz_sets AS sqs, jsonb_to_recordset(sqs.submission_history) AS sh (
        quiz_id text,
        correctness BOOLEAN[],
        filled_text TEXT[],
        is_accepted BOOLEAN,
        correct_text TEXT[], 
        submitted_at timestamp with time zone,
        correct_index INTEGER[],
        selected_index INTEGER[])
	WHERE sqs.deleted_at is NULL
    ORDER BY shuffled_quiz_set_id, quiz_id, created_at DESC
$$;


CREATE OR REPLACE FUNCTION public.migrate_to_lo_submission_answer_fn()
    RETURNS TRIGGER 
    LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO public.lo_submission_answer(
        student_id,
        quiz_id,
        submission_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        student_text_answer,
        correct_text_answer,
        student_index_answer,
        correct_index_answer,
        point,
        is_correct,
        is_accepted,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
    SELECT 
        sh.student_id,
        sh.quiz_id,
        ls.submission_id,
        ls.study_plan_id,
        ls.learning_material_id,
        sh.shuffled_quiz_set_id,
        sh.student_text_answer,
        sh.correct_text_answer,
        sh.student_index_answer,
        sh.correct_index_answer,
        sh.point,
        sh.is_correct,
        sh.is_accepted,
        sh.created_at,
        sh.updated_at,
        sh.deleted_at,
        sh.resource_path
    FROM get_submission_history() AS sh
    JOIN lo_submission ls USING(shuffled_quiz_set_id)
    JOIN quizzes q ON q.external_id = sh.quiz_id
    WHERE ls.deleted_at IS NULL
        AND q.deleted_at IS NULL
        AND sh.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        point = EXCLUDED.point,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at;
RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS migrate_to_lo_submission_answer ON public.shuffled_quiz_sets;
CREATE TRIGGER migrate_to_lo_submission_answer
AFTER UPDATE OF submission_history ON public.shuffled_quiz_sets
FOR EACH ROW
EXECUTE FUNCTION public.migrate_to_lo_submission_answer_fn();