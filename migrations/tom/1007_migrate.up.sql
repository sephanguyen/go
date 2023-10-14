ALTER TABLE ONLY public.messages
  ADD COLUMN IF NOT EXISTS deleted_by TEXT,
  ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
