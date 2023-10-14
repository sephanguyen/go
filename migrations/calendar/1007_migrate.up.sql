ALTER TABLE public.date_type
    ADD COLUMN IF NOT EXISTS display_name text NULL,
    ADD COLUMN IF NOT EXISTS is_archived boolean NOT NULL DEFAULT false;