CREATE INDEX IF NOT EXISTS conversation_lesson_lesson_id_idx ON public.conversation_lesson (lesson_id);

CREATE INDEX IF NOT EXISTS message_created_at__idx ON public.messages (created_at);