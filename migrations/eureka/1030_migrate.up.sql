DROP INDEX IF EXISTS study_plan_content_structure_idx;
CREATE UNIQUE INDEX IF NOT EXISTS study_plan_content_structure_idx
  ON public.study_plan_items (study_plan_id, content_structure_flatten);
