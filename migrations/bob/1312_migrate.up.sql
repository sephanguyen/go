ALTER TABLE public.lesson_room_states ADD COLUMN IF NOT EXISTS recording JSONB DEFAULT '{}'::jsonb;
