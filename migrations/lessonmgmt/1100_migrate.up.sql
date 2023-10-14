ALTER TABLE IF EXISTS public.lesson_room_states
    ADD COLUMN IF NOT EXISTS session_time timestamptz NULL;

ALTER TABLE IF EXISTS public.live_room_state
    ADD COLUMN IF NOT EXISTS session_time timestamptz NULL;