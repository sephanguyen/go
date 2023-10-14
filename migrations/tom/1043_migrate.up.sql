CREATE TABLE IF NOT EXISTS public.private_conversation_lesson (
    conversation_id text,
    lesson_id text NOT NULL,
    flatten_user_ids text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    latest_start_time timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,
    CONSTRAINT private_conversation_lesson_pkey PRIMARY KEY (conversation_id),
    CONSTRAINT private_conversation_lesson_fk FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id),
    CONSTRAINT private_conversation_lesson_unique_lessonid_flatten_user_ids UNIQUE(lesson_id, flatten_user_ids)
);

CREATE POLICY rls_private_conversation_lesson ON "private_conversation_lesson"
    USING (permission_check(resource_path, 'private_conversation_lesson'))
    WITH CHECK (permission_check(resource_path, 'private_conversation_lesson'));

CREATE POLICY rls_private_conversation_lesson_restrictive ON "private_conversation_lesson"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'private_conversation_lesson'))
    WITH CHECK (permission_check(resource_path, 'private_conversation_lesson'));
    
ALTER TABLE "private_conversation_lesson" ENABLE ROW LEVEL security;
ALTER TABLE "private_conversation_lesson" FORCE ROW LEVEL security;

ALTER TABLE ONLY public.conversations DROP CONSTRAINT IF EXISTS conversation_type_check;
ALTER TABLE public.conversations
    ADD CONSTRAINT conversation_type_check CHECK (conversation_type = ANY (ARRAY[
                'CONVERSATION_CLASS'::text,
                'CONVERSATION_QUESTION'::text,
                'CONVERSATION_COACH'::text,
                'CONVERSATION_LESSON'::text,
                'CONVERSATION_STUDENT'::text,
                'CONVERSATION_PARENT'::text,
                'CONVERSATION_LESSON_PRIVATE'::text
            ]));

