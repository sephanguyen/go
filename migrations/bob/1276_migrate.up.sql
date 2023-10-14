ALTER TABLE IF EXISTS public.date_info
    ADD COLUMN IF NOT EXISTS "time_zone" text default current_setting('TIMEZONE');
