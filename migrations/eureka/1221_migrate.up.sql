
DROP TRIGGER IF EXISTS migrate_to_flash_card_submission ON public.shuffled_quiz_sets; 
DROP FUNCTION IF EXISTS migrate_to_flash_card_submission_fn(); --drop old trigger and function to combine

CREATE OR REPLACE FUNCTION public.migrate_to_flash_card_submission_and_flash_card_submission_answer_fn()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    _submission_id text;
BEGIN
    IF EXISTS(
      SELECT 1 FROM flash_card fc WHERE fc.learning_material_id = NEW.learning_material_id
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
            total_point = EXCLUDED.total_point
        RETURNING submission_id into _submission_id;

	    INSERT INTO flash_card_submission_answer(
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
            is_correct,
            is_accepted,
            point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        ) SELECT 
            NEW.student_id,
            fca.quiz_id,
            _submission_id AS submission_id,
            NEW.study_plan_id       ,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            fca.filled_text AS student_text_answer,
            fca.correct_text AS correct_text_answer,
            fca.selected_index AS student_index_answer,
            fca.correct_index AS correct_index_answer,
	        COALESCE(fca.correctness, '{}')::BOOLEAN[] AS is_correct, -- to avoid not null constraint
            fca.is_accepted AS is_accepted,
            COALESCE(fca.is_accepted::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = fca.quiz_id), 0) AS point, 
            COALESCE(fca.submitted_at, NEW.created_at) AS created_at, -- to avoid not null constraint
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
	    FROM (
            SELECT DISTINCT ON(quiz_id) shuffled_quiz_set_id,
                x.* FROM shuffled_quiz_sets sqs, jsonb_to_recordset(sqs.submission_history)
            AS x (quiz_id TEXT,  
                filled_text text[],
                correct_text TEXT[], 
                selected_index INTEGER[], 
                correct_index INTEGER[],
                correctness BOOLEAN[], 
                submitted_at timestamp with time zone,
                is_accepted BOOLEAN)
            WHERE sqs.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
            ORDER BY quiz_id, submitted_at desc
             -- sort by submitted_at desc to ensure get the lastest answer
        ) fca
	    ON CONFLICT ON CONSTRAINT flash_card_submission_answer_pk DO UPDATE SET
            student_text_answer = EXCLUDED.student_text_answer,
            correct_text_answer = EXCLUDED.correct_text_answer,
            student_index_answer = EXCLUDED.student_index_answer,
            correct_index_answer = EXCLUDED.correct_index_answer,
            is_correct = EXCLUDED.is_correct,
            is_accepted = EXCLUDED.is_accepted,
            point = EXCLUDED.point,
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at;	  
    END IF;
RETURN NULL;
END;   
$$;

DROP TRIGGER IF EXISTS migrate_to_flash_card_submission_and_flash_card_submission_answer 
ON public.shuffled_quiz_sets;
CREATE TRIGGER migrate_to_flash_card_submission_and_flash_card_submission_answer
AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets
FOR EACH ROW
EXECUTE FUNCTION public.migrate_to_flash_card_submission_and_flash_card_submission_answer_fn();	


-- Migrate data into flash_card_submission_answer
INSERT INTO public.flash_card_submission_answer (
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
SELECT  fca.student_id,
        fca.submission_id,
        fca.study_plan_id,
        fca.learning_material_id,
        fca.shuffled_quiz_set_id,
        fca.quiz_id,
        fca.filled_text AS student_text_answer,
       	fca.correct_text AS correct_text_answer,
        fca.selected_index AS student_index_answer,
        fca.correct_index AS correct_index_answer,
	    COALESCE(fca.correctness, '{}')::BOOLEAN[] AS is_correct, -- to avoid not null constraint
        fca.is_accepted AS is_accepted,
        COALESCE(fca.is_accepted::INT*(SELECT point FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = fca.quiz_id), 0) AS point, 
        COALESCE(fca.submitted_at, fca.created_at) AS created_at, -- to avoid not null constraint
        fca.updated_at,
        fca.deleted_at,
        fca.resource_path
FROM (
    (SELECT DISTINCT ON(shuffled_quiz_set_id, quiz_id) shuffled_quiz_set_id,
            x.* FROM shuffled_quiz_sets sqs, jsonb_to_recordset(sqs.submission_history)
        AS x (quiz_id TEXT,  
            filled_text text[],
            correct_text TEXT[], 
            selected_index INTEGER[], 
            correct_index INTEGER[],
            correctness BOOLEAN[], 
            submitted_at timestamp with time zone,
            is_accepted BOOLEAN)
        WHERE sqs.deleted_at is NULL -- to avoid wrong data on staging
        ORDER BY shuffled_quiz_set_id, quiz_id, submitted_at desc
    ) sfc -- join to flash card submission
	JOIN flash_card_submission fcs
		USING (shuffled_quiz_set_id)
) fca
ON CONFLICT ON CONSTRAINT flash_card_submission_answer_pk DO NOTHING;