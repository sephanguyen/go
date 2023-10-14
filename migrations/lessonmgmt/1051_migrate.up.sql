ALTER TABLE ONLY public.lesson_student_subscriptions
    ADD COLUMN IF NOT EXISTS "purchased_slot_total" INTEGER;
