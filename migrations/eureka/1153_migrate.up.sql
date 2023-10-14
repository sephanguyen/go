ALTER TABLE IF EXISTS public.exam_lo_submission_score
    DROP COLUMN IF EXISTS shuffle_quiz_set_id,
    ADD COLUMN IF NOT EXISTS shuffled_quiz_set_id TEXT NOT NULL;
