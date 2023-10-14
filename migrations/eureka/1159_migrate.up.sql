ALTER TABLE IF EXISTS student_submissions
  ALTER COLUMN study_plan_item_id DROP NOT NULL,
  ALTER COLUMN assignment_id DROP NOT NULL;

-- drop the primary key contraint
ALTER TABLE IF EXISTS student_latest_submissions
  DROP CONSTRAINT IF EXISTS student_latest_submissions_pk;

ALTER TABLE IF EXISTS student_latest_submissions
  ALTER COLUMN study_plan_item_id DROP NOT NULL,
  ALTER COLUMN assignment_id DROP NOT NULL,
  DROP CONSTRAINT IF EXISTS student_latest_submissions_uk,
  DROP CONSTRAINT IF EXISTS student_latest_submissions_old_uk,
  -- use UNIQUE instead because PRIMARY KEY enforces NOT NULL
  ADD CONSTRAINT student_latest_submissions_uk UNIQUE (student_id, study_plan_id, learning_material_id),
  ADD CONSTRAINT student_latest_submissions_old_uk UNIQUE (student_id, study_plan_item_id, assignment_id);