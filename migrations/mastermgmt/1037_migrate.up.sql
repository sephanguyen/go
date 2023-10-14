ALTER TABLE ONLY public.grade
  ADD COLUMN IF NOT EXISTS remarks text;