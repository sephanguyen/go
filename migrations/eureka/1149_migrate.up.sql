ALTER TABLE IF EXISTS study_plan_monitors
ADD COLUMN IF NOT EXISTS auto_upserted_at timestamp with time zone;