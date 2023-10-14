ALTER TABLE public.applied_slot ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;

ALTER TABLE public.center_opening_slot ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;

ALTER TABLE public.student_available_slot_master ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;

ALTER TABLE public.teacher_subject ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;

ALTER TABLE public.teacher_available_slot_master ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;

ALTER TABLE public.time_slot ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;

ALTER TABLE public.job_schedule_status ADD COLUMN IF NOT EXISTS run_time_id VARCHAR(255) NULL;
