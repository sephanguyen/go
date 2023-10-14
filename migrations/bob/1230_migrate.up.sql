ALTER TABLE public.school_info 
    DROP COLUMN IF EXISTS prefecture,
    DROP COLUMN IF EXISTS city,
    ADD COLUMN IF NOT EXISTS school_level_id TEXT NOT NULL,
    ADD COLUMN IF NOT EXISTS address TEXT,
    DROP CONSTRAINT IF EXISTS school_info__school_level_id__fk,
    ADD CONSTRAINT school_info__school_level_id__fk FOREIGN KEY (school_level_id) REFERENCES public.school_level(school_level_id);
