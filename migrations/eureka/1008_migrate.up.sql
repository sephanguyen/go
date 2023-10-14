ALTER TABLE ONLY study_plan_items ADD COLUMN IF NOT EXISTS completed_at timestamptz;
