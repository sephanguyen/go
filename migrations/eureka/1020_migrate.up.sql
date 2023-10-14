ALTER TABLE IF EXISTS public.student_submission_grades
  ADD COLUMN IF NOT EXISTS "status" TEXT,
  ADD COLUMN IF NOT EXISTS "editor_id" TEXT;

ALTER TABLE IF EXISTS public.student_submission_grades ALTER COLUMN "grader_id" DROP NOT NULL;
