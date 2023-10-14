ALTER TABLE IF EXISTS public.scheduler_patterns
    ALTER COLUMN end_time DROP NOT NULL;

ALTER TABLE IF EXISTS public.scheduler_items
    ALTER COLUMN end_time DROP NOT NULL;
