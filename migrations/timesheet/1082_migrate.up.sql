ALTER TABLE ONLY public.lessons
    ADD COLUMN IF NOT EXISTS "preparation_time" INTEGER,
    ADD COLUMN IF NOT EXISTS "break_time" INTEGER;
