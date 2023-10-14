ALTER TABLE ONLY public.lessons
    ADD COLUMN IF NOT EXISTS "lesson_capacity" INTEGER;
