ALTER TABLE public.student_package_order ADD COLUMN IF NOT EXISTS from_student_package_order_id text;

ALTER TABLE public.student_package_order ALTER COLUMN start_at DROP NOT NULL;
ALTER TABLE public.student_package_order ALTER COLUMN end_at DROP NOT NULL;

ALTER TABLE public.student_package_order ADD COLUMN IF NOT EXISTS is_executed_by_cronjob boolean NOT NULL DEFAULT false;
ALTER TABLE public.student_package_order ADD COLUMN IF NOT EXISTS executed_error text;
