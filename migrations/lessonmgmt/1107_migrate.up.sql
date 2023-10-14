ALTER TABLE IF EXISTS public.live_lesson_conversation
    ADD COLUMN IF NOT EXISTS "conversation_type" TEXT NOT NULL
;