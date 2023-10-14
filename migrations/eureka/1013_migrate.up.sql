-- ALTER TABLE public.student_submission_grades
--   ALTER COLUMN grade TYPE NUMERIC(10, 2);

ALTER TABLE public.student_submission_grades
  DROP COLUMN IF EXISTS grade;

ALTER TABLE public.student_submission_grades
  ADD COLUMN IF NOT EXISTS grade NUMERIC(10, 2);
