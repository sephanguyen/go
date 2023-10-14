ALTER TABLE ONLY public.staff ADD COLUMN IF NOT EXISTS auto_create_timesheet boolean DEFAULT false;
