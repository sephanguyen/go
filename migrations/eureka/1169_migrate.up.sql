ALTER TABLE IF EXISTS public.student_latest_submissions REPLICA IDENTITY FULL;

-- replace old trigger with BEFORE INSERT behaviour
CREATE OR REPLACE FUNCTION public.trigger_student_latest_submissions_fill_new_identity_fn()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE sp_id TEXT;
BEGIN
  IF NEW.study_plan_item_id IS NULL THEN
    RETURN NEW;
  END IF;

  SELECT nid.study_plan_id INTO sp_id FROM public.retrieve_study_plan_identity(ARRAY[NEW.study_plan_item_id]) as nid;
  NEW.study_plan_id = sp_id;
  NEW.learning_material_id = NEW.assignment_id;

  RETURN NEW;
END;
$function$
;

DROP TRIGGER IF EXISTS fill_new_identity ON student_latest_submissions;

CREATE TRIGGER fill_new_identity BEFORE INSERT ON public.student_latest_submissions
FOR EACH ROW EXECUTE FUNCTION trigger_student_latest_submissions_fill_new_identity_fn();

-- Fix missing study_plan_id for student's adhoc task assignment
UPDATE student_latest_submissions sls SET
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
  sls.study_plan_id IS NULL;

UPDATE student_submissions ss SET
  study_plan_id =(
    SELECT
      COALESCE(sp.master_study_plan_id, sp.study_plan_id)
    FROM
      study_plan_items spi
      JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
    WHERE
      spi.study_plan_item_id = ss.study_plan_item_id
  ),
  learning_material_id = ss.assignment_id
WHERE
  ss.study_plan_id IS NULL;


ALTER TABLE IF EXISTS student_latest_submissions
  DROP CONSTRAINT IF EXISTS student_latest_submissions_uk,
  ADD CONSTRAINT student_latest_submissions_pk PRIMARY KEY (student_id, study_plan_id, learning_material_id);
ALTER TABLE IF EXISTS public.student_latest_submissions REPLICA IDENTITY DEFAULT;
