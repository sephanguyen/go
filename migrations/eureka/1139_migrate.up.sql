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
      COALESCE(sp.master_study_plan_id, sp.study_plan_id)
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

-- migration student_latest_submissions to fix null study_plan_id
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
  learning_material_id = sls.assignment_id
WHERE
  sls.study_plan_id IS NULL;

-- fix missing migration student_submissions of 1135
UPDATE student_submissions ss SET
  study_plan_id =(
    SELECT
      sp.master_study_plan_id
    FROM
      study_plan_items spi
      JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
    WHERE
      spi.study_plan_item_id = ss.study_plan_item_id
  ),
  learning_material_id = ss.assignment_id;
