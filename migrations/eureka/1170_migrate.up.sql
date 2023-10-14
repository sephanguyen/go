ALTER TABLE IF EXISTS flashcard_progressions
  ALTER COLUMN study_plan_item_id DROP NOT NULL,
  ALTER COLUMN lo_id DROP NOT NULL;
