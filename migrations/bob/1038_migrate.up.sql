ALTER TABLE public.lessons
  ADD COLUMN IF NOT EXISTS room_id TEXT;
