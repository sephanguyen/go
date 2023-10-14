ALTER TABLE IF EXISTS student_submissions
    ADD COLUMN IF NOT EXISTS deleted_by TEXT;

ALTER TABLE IF EXISTS student_latest_submissions
    ADD COLUMN IF NOT EXISTS deleted_by TEXT;
