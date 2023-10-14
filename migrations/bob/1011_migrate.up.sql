ALTER TABLE public.lessons DROP CONSTRAINT IF EXISTS lessons_pk CASCADE;

ALTER TABLE ONLY public.lessons
    ADD CONSTRAINT lessons_pk PRIMARY KEY (lesson_id);
