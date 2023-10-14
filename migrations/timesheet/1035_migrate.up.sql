ALTER TABLE IF EXISTS public.auto_create_flag_activity_log DROP COLUMN IF EXISTS end_time;

ALTER TABLE IF EXISTS public.auto_create_flag_activity_log
    RENAME start_time TO change_time;