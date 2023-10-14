ALTER TABLE public.lesson_room_states ADD COLUMN IF NOT EXISTS whiteboard_zoom_state JSONB DEFAULT '{}'::jsonb;
