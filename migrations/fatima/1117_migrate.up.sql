ALTER TABLE public.order
    ADD COLUMN IF NOT EXISTS loa_start_date timestamptz DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS loa_end_date timestamptz DEFAULT NULL;