CREATE OR REPLACE FUNCTION update_question_hierarchy_quiz_sets_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN 
    UPDATE public.quiz_sets
    SET question_hierarchy = (
        CASE WHEN ARRAY_LENGTH(quiz_external_ids, 1) IS NULL
        THEN ARRAY[]::JSONB[]
        ELSE
        (
            SELECT ARRAY_AGG(
                TO_JSONB(qei)
            ) FROM (
                SELECT UNNEST(quiz_external_ids) as id, 'QUESTION' as type
            ) qei
        )
        END
    )
    WHERE quiz_set_id=NEW.quiz_set_id
    AND NEW.question_hierarchy IS NULL
    AND NOT EXISTS(
        SELECT 1 FROM flash_card WHERE learning_material_id=NEW.lo_id AND deleted_at IS NULL
    );
RETURN NULL;
END;
$$ LANGUAGE plpgsql;
