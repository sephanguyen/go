ALTER TABLE IF EXISTS public.student_latest_submissions
  ADD COLUMN IF NOT EXISTS study_plan_id TEXT,
  ADD COLUMN IF NOT EXISTS learning_material_id TEXT;

-- trigger
CREATE OR REPLACE FUNCTION public.trigger_student_latest_submissions_fill_new_identity_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
BEGIN
UPDATE
  student_latest_submissions sls
SET
  study_plan_id =(
    SELECT
      sp.master_study_plan_id
    FROM
      study_plan_items spi
      JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
    WHERE
      spi.study_plan_item_id = sls.study_plan_item_id
  ),
  learning_material_id = sls.assignment_id
WHERE
  study_plan_item_id = NEW.study_plan_item_id;
RETURN NULL;
END;
$function$;

DROP TRIGGER IF EXISTS fill_new_identity ON public.student_latest_submissions;
CREATE TRIGGER fill_new_identity AFTER INSERT ON public.student_latest_submissions
  FOR EACH ROW EXECUTE FUNCTION trigger_student_latest_submissions_fill_new_identity_fn();

-- migration
UPDATE student_latest_submissions sls SET
  study_plan_id =(
    SELECT
      sp.master_study_plan_id
    FROM
      study_plan_items spi
      JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
    WHERE
      spi.study_plan_item_id = sls.study_plan_item_id
  ),
  learning_material_id = sls.assignment_id;