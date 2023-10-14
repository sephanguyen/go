ALTER TABLE public.e2e_instances
    ADD COLUMN IF NOT EXISTS message TEXT DEFAULT '';
