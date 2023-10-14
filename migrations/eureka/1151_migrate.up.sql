ALTER TABLE IF EXISTS public.shuffled_quiz_sets ADD COLUMN IF NOT EXISTS question_hierarchy JSONB[] DEFAULT ARRAY[]::JSONB[];

CREATE OR REPLACE FUNCTION update_question_hierarchy_shuffled_quiz_sets_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN 
    UPDATE public.shuffled_quiz_sets
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
    WHERE shuffled_quiz_set_id=NEW.shuffled_quiz_set_id;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS question_hierarchy ON public.shuffled_quiz_sets;
CREATE TRIGGER question_hierarchy AFTER INSERT ON public.shuffled_quiz_sets FOR EACH ROW EXECUTE FUNCTION public.update_question_hierarchy_shuffled_quiz_sets_fn();
