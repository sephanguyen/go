ALTER TABLE ONLY public.conversation_members DROP CONSTRAINT IF EXISTS conversation_statuses_status_check;
ALTER TABLE ONLY public.conversation_members
    ADD CONSTRAINT conversation_statuses_status_check CHECK ((status = ANY ('{CONVERSATION_STATUS_ACTIVE,CONVERSATION_STATUS_INACTIVE,CONVERSATION_STATUS_MUTED}'::text[])))
