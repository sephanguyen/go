ALTER TABLE public.activity_logs ADD COLUMN IF NOT EXISTS finished_at timestamptz default NULL;
