-- modify table
ALTER TABLE IF EXISTS public.student_event_logs
    ADD COLUMN IF NOT EXISTS study_plan_id TEXT,
    ADD COLUMN IF NOT EXISTS learning_material_id TEXT;
-- trigger

--drop previous trigger
DROP TRIGGER IF EXISTS fill_study_plan_item_id ON public.student_event_logs;
-- drop func
DROP FUNCTION IF EXISTS fill_study_plan_item_id_fn;
--trigger 
CREATE OR REPLACE FUNCTION public.trigger_student_event_logs_fill_new_identity_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
BEGIN IF new.event_type = ANY(ARRAY[
    'study_guide_finished',
    'video_finished',
    'learning_objective',
    'quiz_answer_selected'
    ]) THEN
UPDATE
    student_event_logs sel
SET
    study_plan_item_id = (new.payload ->> 'study_plan_item_id'),
    study_plan_id =(
        SELECT
        COALESCE(sp.master_study_plan_id, sp.study_plan_id)
        FROM
        study_plan_items spi
        JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
        WHERE
        spi.study_plan_item_id = new.payload ->> 'study_plan_item_id'
    ),
    learning_material_id = new.payload->>'lo_id'
WHERE
    student_event_log_id = new.student_event_log_id;
END IF;
RETURN NULL;
END;
$function$;

DROP TRIGGER IF EXISTS fill_new_identity ON public.student_event_logs;
CREATE TRIGGER fill_new_identity AFTER INSERT ON public.student_event_logs
    FOR EACH ROW EXECUTE FUNCTION trigger_student_event_logs_fill_new_identity_fn();
