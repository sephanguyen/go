CREATE OR REPLACE FUNCTION update_question_hierachy_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN 
    UPDATE public.quiz_sets
    SET question_hierachy = (
        SELECT JSON_OBJECT_AGG(qei.key, qei.value)
        FROM (SELECT UNNEST(quiz_external_ids) as key, null as value) qei
    )
    WHERE quiz_set_id=NEW.quiz_set_id;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS question_hierachy ON public.quiz_sets;
CREATE TRIGGER question_hierachy AFTER INSERT ON public.quiz_sets FOR EACH ROW EXECUTE FUNCTION public.update_question_hierachy_fn();
