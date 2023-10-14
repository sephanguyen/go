DROP INDEX IF EXISTS conversation_ids_idx;
DROP INDEX IF EXISTS conversation_members_conversation_id_idx;

CREATE INDEX IF NOT EXISTS conversation_members_conversation_id_idx ON public.conversation_members USING btree (conversation_id);
