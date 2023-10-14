ALTER TABLE public.job_schedule_status ADD COLUMN IF NOT EXISTS scheduling_jobs_id VARCHAR(255) NULL;
