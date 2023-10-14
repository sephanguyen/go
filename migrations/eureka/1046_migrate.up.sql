ALTER TABLE IF EXISTS public.student_study_plans
  ADD COLUMN IF NOT EXISTS master_study_plan_id TEXT;

ALTER TABLE IF EXISTS public.student_study_plans
  ADD CONSTRAINT student_master_study_plan UNIQUE (student_id, master_study_plan_id);

CREATE OR REPLACE FUNCTION update_master_study_plan_id_on_student_study_plan_created_fn()
RETURNS TRIGGER
AS $$
  BEGIN
    UPDATE student_study_plans
    SET master_study_plan_id = (SELECT master_study_plan_id FROM study_plans WHERE study_plan_id=NEW.study_plan_id)
    WHERE study_plan_id = NEW.study_plan_id;
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_master_study_plan_student_study_plan ON public.student_study_plans;
CREATE TRIGGER update_master_study_plan_student_study_plan
  AFTER INSERT
  ON student_study_plans
  FOR EACH ROW
  EXECUTE FUNCTION update_master_study_plan_id_on_student_study_plan_created_fn();

UPDATE student_study_plans ssp SET master_study_plan_id = sp.master_study_plan_id
FROM study_plans sp
WHERE sp.study_plan_id = ssp.study_plan_id;
