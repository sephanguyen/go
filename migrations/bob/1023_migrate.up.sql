ALTER TABLE ONLY public.lessons ADD COLUMN IF NOT EXISTS control_settings JSONB NULL;
