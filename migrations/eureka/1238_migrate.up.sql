ALTER TABLE IF EXISTS public.lo_progression_answer
    ADD COLUMN IF NOT EXISTS submitted_keys_answer text[];