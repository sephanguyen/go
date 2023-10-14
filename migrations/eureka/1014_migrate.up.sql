ALTER TABLE public.student_submissions
  ADD COLUMN IF NOT EXISTS editor_id TEXT;
