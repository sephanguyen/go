CREATE TABLE IF NOT EXISTS public.conversation_students (
    id text,
    conversation_id text not null,
    student_id text not null,
    conversation_type text not null,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    deleted_at timestamp with time zone,
    CONSTRAINT conversation_students_pk PRIMARY KEY (id),
    CONSTRAINT conversation_students_conversations_fk FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id)
);

ALTER TABLE ONLY public.conversation_students DROP CONSTRAINT IF EXISTS conversation_type;
ALTER TABLE public.conversation_students
    ADD CONSTRAINT conversation_type CHECK (conversation_type = ANY (ARRAY[
        'CONVERSATION_STUDENT',
        'CONVERSATION_PARENT'
]::text[]));


CREATE INDEX IF NOT EXISTS conversation_students_conversation_id_idx ON public.conversation_students(conversation_id);
CREATE INDEX IF NOT EXISTS conversation_students_student_id_idx ON public.conversation_students(student_id);


ALTER TABLE ONLY public.conversations
     ADD COLUMN IF NOT EXISTS owner text;

ALTER TABLE ONLY public.conversations DROP CONSTRAINT IF EXISTS conversation_type_check;
ALTER TABLE public.conversations
    ADD CONSTRAINT conversation_type_check CHECK (conversation_type = ANY (ARRAY[
                'CONVERSATION_CLASS'::text,
                'CONVERSATION_QUESTION'::text,
                'CONVERSATION_COACH'::text,
                'CONVERSATION_LESSON'::text,
                'CONVERSATION_STUDENT'::text,
                'CONVERSATION_PARENT'::text
            ]));
