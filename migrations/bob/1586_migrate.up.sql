ALTER TABLE IF EXISTS public.lesson_room_states
    ADD COLUMN IF NOT EXISTS session_time timestamptz NULL;