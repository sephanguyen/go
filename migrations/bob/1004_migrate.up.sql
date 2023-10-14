ALTER TABLE ONLY public.questions
    ADD COLUMN IF NOT EXISTS question_url text NULL;

ALTER TABLE ONLY public.questions
    ADD COLUMN IF NOT EXISTS answers_url text[] NULL;

ALTER TABLE ONLY public.questions
    ADD COLUMN IF NOT EXISTS explanation_url text NULL;

ALTER TABLE ONLY public.questions
    ADD COLUMN IF NOT EXISTS explanation_wrong_answer_url text[] NULL;

ALTER TABLE ONLY public.questions
    ADD COLUMN IF NOT EXISTS rendering_question boolean NULL;
