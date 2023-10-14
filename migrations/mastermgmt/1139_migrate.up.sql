ALTER TABLE ONLY public.configuration_group ADD COLUMN IF NOT EXISTS is_released boolean DEFAULT false;
