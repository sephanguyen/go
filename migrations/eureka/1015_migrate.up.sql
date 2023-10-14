DROP INDEX IF EXISTS copy_study_plan_item_id_idx;
CREATE INDEX IF NOT EXISTS copy_study_plan_item_id ON public.study_plan_items (copy_study_plan_item_id);

DROP INDEX IF EXISTS master_study_plan_item_id_idx;
CREATE INDEX IF NOT EXISTS master_study_plan_item_id_idx ON public.study_plans (master_study_plan_id);
