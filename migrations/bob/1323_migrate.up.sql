ALTER TABLE public.media
    ADD COLUMN IF NOT EXISTS file_size_bytes bigint DEFAULT 0,
    ADD COLUMN IF NOT EXISTS duration_seconds integer DEFAULT 0;
