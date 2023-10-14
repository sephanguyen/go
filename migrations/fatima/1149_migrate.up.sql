ALTER TABLE public.upcoming_student_package ADD COLUMN IF NOT EXISTS is_executed_by_cronjob boolean NOT NULL DEFAULT false;
ALTER TABLE public.upcoming_student_package ADD COLUMN IF NOT EXISTS executed_error text;
ALTER TABLE public.upcoming_student_course ADD COLUMN IF NOT EXISTS is_executed_by_cronjob boolean NOT NULL DEFAULT false;
ALTER TABLE public.upcoming_student_course ADD COLUMN IF NOT EXISTS executed_error text;

