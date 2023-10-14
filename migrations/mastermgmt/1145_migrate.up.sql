ALTER TABLE ONLY public.configuration_group ADD COLUMN IF NOT EXISTS launching_date date;
ALTER TABLE public.configuration_group DROP COLUMN IF EXISTS is_released;
