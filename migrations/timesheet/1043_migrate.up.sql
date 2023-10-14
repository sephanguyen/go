-- drop column `is_valid` on `public.timesheet`
ALTER TABLE IF EXISTS public.timesheet
    DROP COLUMN IF EXISTS is_valid;

-- add column `flag_on` on `public.timesheet_lesson_hours`
ALTER TABLE IF EXISTS public.timesheet_lesson_hours
    ADD COLUMN IF NOT EXISTS flag_on BOOLEAN NOT NULL DEFAULT FALSE;