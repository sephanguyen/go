ALTER TABLE IF EXISTS public.conversation_statuses DROP CONSTRAINT IF EXISTS conversation_statuses__conversation_id__fk;
ALTER TABLE IF EXISTS public.messages DROP CONSTRAINT IF EXISTS messages__conversation_id__fk;


ALTER TABLE IF EXISTS public.conversations ALTER COLUMN conversation_id TYPE text USING conversation_id::text;
ALTER TABLE IF EXISTS public.conversation_statuses ALTER COLUMN conversation_id TYPE text USING conversation_id::text;
ALTER TABLE IF EXISTS public.conversation_statuses ALTER COLUMN conversation_statuses_id TYPE text USING conversation_statuses_id::text;
ALTER TABLE IF EXISTS public.messages ALTER COLUMN message_id TYPE text USING message_id::text;
ALTER TABLE IF EXISTS public.messages ALTER COLUMN conversation_id TYPE text USING conversation_id::text;

ALTER TABLE IF EXISTS public.conversation_statuses ADD CONSTRAINT conversation_statuses__conversation_id__fk FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id);
ALTER TABLE IF EXISTS public.messages ADD CONSTRAINT messages__conversation_id__fk FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id);
