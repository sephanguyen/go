CREATE OR REPLACE FUNCTION public.trigger_student_submissions_fill_new_identity_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
DECLARE current_study_plan_id text;
        current_learning_material_id text;
        current_student_id text;
        current_resource_path text;
        tmp record;
BEGIN
  current_study_plan_id = (
  SELECT
    COALESCE(sp.master_study_plan_id, sp.study_plan_id)
  FROM
    study_plan_items spi
  JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
  WHERE
    spi.study_plan_item_id = new.study_plan_item_id
  );
  current_learning_material_id = new.assignment_id;
  current_student_id = new.student_id;
  current_resource_path = new.resource_path;
  IF TG_OP = 'INSERT' THEN 
    UPDATE
    student_submissions ss
    SET
    study_plan_id = current_study_plan_id,
    learning_material_id = current_learning_material_id
    WHERE
    study_plan_item_id = NEW.study_plan_item_id;
  ELSE 
     -- if TG_OP != 'INSERT', in some cases we don't have new.study_plan_item_id so we need to use student_submission_id for query to get identity  
    IF (current_study_plan_id IS NULL) AND (new.student_submission_id is not null) THEN
        SELECT
        ss.study_plan_id,
        ss.learning_material_id,
        ss.student_id,
        ss.resource_path
        FROM
        student_submissions ss
        WHERE
        ss.student_submission_id = new.student_submission_id
        INTO tmp;
    current_study_plan_id = tmp.study_plan_id;
    current_learning_material_id = tmp.learning_material_id;
    current_student_id = tmp.student_id;
    current_resource_path = tmp.resource_path;
    END IF;
  END IF ;
  
-- this is the case for assignment submission and task assignment submission (don't care status SUBMISSION_STATUS_RETURNED )
  IF (new.status= 'SUBMISSION_STATUS_RETURNED') OR EXISTS(select 1 from task_assignment where learning_material_id = current_learning_material_id) THEN
	  call upsert_highest_score(current_study_plan_id, current_learning_material_id, current_student_id, current_resource_path);
  END IF;
RETURN NULL;
END;
$function$;

DROP TRIGGER IF EXISTS fill_new_identity ON public.student_submissions;
CREATE TRIGGER fill_new_identity AFTER INSERT OR UPDATE ON public.student_submissions
  FOR EACH ROW EXECUTE FUNCTION trigger_student_submissions_fill_new_identity_fn();
