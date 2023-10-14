ALTER TABLE ONLY public.exam_lo_submission_answer
    ADD COLUMN IF NOT EXISTS submitted_keys_answer text[],
    ADD COLUMN IF NOT EXISTS correct_keys_answer text[];
