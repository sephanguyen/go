ALTER TABLE ONLY public.lesson_room_states
    ADD COLUMN IF NOT EXISTS agora_room_id TEXT;

ALTER TABLE ONLY public.lesson_room_states
    ADD COLUMN IF NOT EXISTS streaming_attendees TEXT[] NOT NULL DEFAULT '{}'::TEXT[];

ALTER TABLE ONLY public.lesson_room_states
    ADD COLUMN IF NOT EXISTS ended_at timestamp with time zone;
