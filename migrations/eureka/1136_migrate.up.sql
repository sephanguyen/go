ALTER TABLE IF EXISTS shuffled_quiz_sets 
    ADD COLUMN IF NOT EXISTS study_plan_id TEXT NULL,
    ADD COLUMN IF NOT EXISTS learning_material_id TEXT NULL;

CREATE OR REPLACE FUNCTION update_study_plan_item_identity_for_shuffled_quiz_set_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN 
    UPDATE public.shuffled_quiz_sets sqs
    SET study_plan_id =  COALESCE(sp.master_study_plan_id, sp.study_plan_id),
        learning_material_id = CASE
            WHEN content_structure ->> 'lo_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'lo_id'
            WHEN content_structure ->> 'assignment_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'assignment_id'
            ELSE NULL 
            END 
    FROM public.study_plan_items spi
    JOIN public.study_plans sp
    USING(study_plan_id)
    WHERE sqs.study_plan_item_id = new.study_plan_item_id AND sqs.study_plan_item_id = spi.study_plan_item_id;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_study_plan_item_identity_for_shuffled_quiz_set_fn ON public.shuffled_quiz_sets;

DROP TRIGGER IF EXISTS update_study_plan_item_identity_for_shuffled_quiz_set ON public.shuffled_quiz_sets;
CREATE TRIGGER update_study_plan_item_identity_for_shuffled_quiz_set AFTER INSERT ON shuffled_quiz_sets FOR EACH ROW EXECUTE FUNCTION update_study_plan_item_identity_for_shuffled_quiz_set_fn();

UPDATE public.shuffled_quiz_sets sqs
    SET study_plan_id =  COALESCE(sp.master_study_plan_id, sp.study_plan_id),
        learning_material_id = CASE
            WHEN content_structure ->> 'lo_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'lo_id'
            WHEN content_structure ->> 'assignment_id' != ANY(ARRAY['', NULL]) THEN content_structure ->> 'assignment_id'
            ELSE NULL
            END 
    FROM public.study_plan_items spi
    JOIN public.study_plans sp
    USING(study_plan_id)
WHERE sqs.study_plan_item_id = spi.study_plan_item_id;