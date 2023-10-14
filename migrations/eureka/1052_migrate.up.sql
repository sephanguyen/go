DROP INDEX IF EXISTS copy_study_plan_item_id;
CREATE INDEX IF NOT EXISTS copy_study_plan_item_id_idx ON public.study_plan_items (copy_study_plan_item_id);