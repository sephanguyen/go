ALTER TABLE shuffled_quiz_sets
	ADD COLUMN IF NOT EXISTS student_id TEXT NOT NULL,
  	ADD COLUMN IF NOT EXISTS study_plan_item_id TEXT,
	ADD COLUMN IF NOT EXISTS total_correctness INTEGER DEFAULT 0 NOT NULL,
	ADD COLUMN IF NOT EXISTS submission_history JSONB DEFAULT '[]'::JSONB NOT NULL;

ALTER TABLE shuffled_quiz_sets DROP CONSTRAINT IF EXISTS student_id_fk;
ALTER TABLE shuffled_quiz_sets ADD CONSTRAINT student_id_fk FOREIGN KEY (student_id) REFERENCES students(student_id);
