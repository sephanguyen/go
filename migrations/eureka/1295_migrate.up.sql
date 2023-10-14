CREATE OR REPLACE FUNCTION public.migrate_to_lo_submission_and_answer_fnc()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $FUNCTION$
BEGIN
  -- insert to lo submission first
    IF EXISTS (
        SELECT 1 
        FROM public.learning_objective LO 
        WHERE LO.learning_material_id = NEW.learning_material_id and NEW.submission_history::text != '[]'::text
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
            COALESCE(
                (
                    SELECT SUM(point) 
                    FROM public.quizzes 
                    WHERE quizzes.deleted_at IS NULL AND quizzes.external_id = ANY(SELECT unnest(quiz_external_ids) FROM quiz_sets qs WHERE 
                    qs.quiz_set_id = new.original_quiz_set_id)
                    ), 0),
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point;

  -- continue insert to lo answer 
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
        submitted_keys_answer,
        correct_keys_answer,
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
        sh.submitted_keys_answer,
        sh.correct_keys_answer,
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
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer,
        point = EXCLUDED.point,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at;
    END IF;
  
RETURN NULL;
END;
$FUNCTION$;
