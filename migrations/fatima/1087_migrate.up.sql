ALTER TABLE ONLY public.grade
  ADD COLUMN IF NOT EXISTS sequence integer,
  ADD CONSTRAINT grade__sequence__unique UNIQUE (sequence, resource_path);
