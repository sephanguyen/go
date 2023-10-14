ALTER TABLE IF EXISTS ONLY public.conversation_students
  ADD COLUMN IF NOT EXISTS search_index_time timestamp with time zone NULL