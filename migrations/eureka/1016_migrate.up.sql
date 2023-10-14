ALTER TABLE public.assignments
    ADD COLUMN IF NOT EXISTS display_order INTEGER DEFAULT 0;
