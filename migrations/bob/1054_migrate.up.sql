ALTER TABLE quizzes 
	ADD COLUMN IF NOT EXISTS lo_id TEXT;

ALTER TABLE quizzes
	DROP CONSTRAINT IF EXISTS lo_id_fk;

ALTER TABLE quizzes
	ADD CONSTRAINT lo_id_fk FOREIGN KEY (lo_id) REFERENCES learning_objectives(lo_id);