ALTER TABLE ONLY public.conversation_lesson
    DROP CONSTRAINT IF EXISTS conversation_lesson_pk;

-- make lesson_id pk, delete all null record
DELETE FROM public.conversation_lesson 
    WHERE lesson_id IS NULL;

-- lesson_id is used more, make it pk
ALTER TABLE ONLY public.conversation_lesson
    ADD CONSTRAINT conversation_lesson_pk PRIMARY KEY (lesson_id);

ALTER TABLE ONLY public.conversation_lesson
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone NULL,
    ADD COLUMN IF NOT EXISTS latest_start_time timestamp with time zone NULL,
    ADD COLUMN IF NOT EXISTS latest_call_id text NULL;

ALTER TABLE ONLY public.conversation_lesson
    DROP CONSTRAINT IF EXISTS conversation_lesson__conversation_id__fk;

ALTER TABLE ONLY public.conversation_lesson
    ADD CONSTRAINT conversation_lesson__conversation_id__fk FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id) NOT VALID;

-- remove orphan record
DELETE FROM public.conversation_lesson cl
WHERE NOT EXISTS (
   SELECT 1 FROM public.conversations c 
   WHERE  c.conversation_id = cl.conversation_id 
);
ALTER TABLE public.conversation_lesson VALIDATE CONSTRAINT conversation_lesson__conversation_id__fk;
-- https://dba.stackexchange.com/questions/158162/deleting-would-be-orphans-when-creating-fk-in-postgres


