ALTER TABLE IF EXISTS student_event_logs ADD COLUMN IF NOT EXISTS study_plan_item_id TEXT NULL;;
-- insert trigger
CREATE OR REPLACE FUNCTION public.fill_study_plan_item_id_fn()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$ 
BEGIN IF new.event_type = ANY(ARRAY[
  'study_guide_finished',
  'video_finished',
  'learning_objective',
  'quiz_answer_selected'
  ]) THEN
UPDATE
    student_event_logs
SET
    study_plan_item_id = (new.payload ->> 'study_plan_item_id')
WHERE
    student_event_log_id = new.student_event_log_id;
END IF;

RETURN NULL;
END;
$function$
;
DROP TRIGGER IF EXISTS fill_study_plan_item_id ON public.student_event_logs;
CREATE TRIGGER fill_study_plan_item_id AFTER INSERT ON public.student_event_logs FOR EACH ROW EXECUTE FUNCTION fill_study_plan_item_id_fn();