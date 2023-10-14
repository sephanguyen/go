DROP TRIGGER IF EXISTS question_hierachy ON public.quiz_sets;
DROP FUNCTION IF EXISTS update_question_hierachy_fn();

ALTER TABLE IF EXISTS public.quiz_sets DROP COLUMN IF EXISTS question_hierachy;

ALTER TABLE IF EXISTS public.quiz_sets ADD COLUMN IF NOT EXISTS question_hierarchy JSONB[] DEFAULT ARRAY[]::JSONB[];

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
    WHERE quiz_set_id=NEW.quiz_set_id;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS question_hierarchy ON public.quiz_sets;
CREATE TRIGGER question_hierarchy AFTER INSERT ON public.quiz_sets FOR EACH ROW EXECUTE FUNCTION public.update_question_hierarchy_quiz_sets_fn();