ALTER TABLE public.student_questions
  ADD COLUMN IF NOT EXISTS grade smallint,
  ADD COLUMN IF NOT EXISTS subject TEXT,
  ALTER COLUMN quiz_id DROP NOT NULL;
