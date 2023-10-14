ALTER TABLE IF EXISTS public.brands
    ADD COLUMN IF NOT EXISTS "time_zone" text,
    ADD COLUMN IF NOT EXISTS "academic_year_beginning" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    ADD COLUMN IF NOT EXISTS "academic_year_end" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    ADD COLUMN IF NOT EXISTS "scheduler_release_status" text;

ALTER TABLE IF EXISTS public.centers
    ADD COLUMN IF NOT EXISTS "time_zone" text;
