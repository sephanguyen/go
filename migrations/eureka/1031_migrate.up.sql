ALTER TABLE IF EXISTS assign_study_plan_tasks
  ADD COLUMN IF NOT EXISTS error_detail TEXT DEFAULT NULL;