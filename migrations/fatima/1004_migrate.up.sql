ALTER TABLE public.student_packages
  ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.packages
  ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

