CREATE TABLE IF NOT EXISTS public.conversation_lesson (
    conversation_id text,
    lesson_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS public.conversation_class (
    conversation_id text,
    class_id integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS public.conversation_question (
    conversation_id text,
    student_question_id text,
    tutor_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS public.conversation_coach (
    conversation_id text,
    coach_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

ALTER TABLE ONLY public.conversations
     ADD COLUMN IF NOT EXISTS conversation_type text;


ALTER TABLE ONLY public.conversation_question DROP CONSTRAINT IF EXISTS conversation_question_pk;
ALTER TABLE ONLY public.conversation_question
    ADD CONSTRAINT conversation_question_pk PRIMARY KEY (conversation_id);


ALTER TABLE ONLY public.conversation_coach DROP CONSTRAINT IF EXISTS conversation_coach_pk;
ALTER TABLE ONLY public.conversation_coach
    ADD CONSTRAINT conversation_coach_pk PRIMARY KEY (conversation_id);


ALTER TABLE ONLY public.conversation_class DROP CONSTRAINT IF EXISTS conversation_class_pk;
ALTER TABLE ONLY public.conversation_class
    ADD CONSTRAINT conversation_class_pk PRIMARY KEY (conversation_id);


ALTER TABLE ONLY public.conversation_lesson DROP CONSTRAINT IF EXISTS conversation_lesson_pk;
ALTER TABLE ONLY public.conversation_lesson
    ADD CONSTRAINT conversation_lesson_pk PRIMARY KEY (conversation_id);

ALTER TABLE ONLY public.conversations DROP CONSTRAINT IF EXISTS conversation_type_check;
ALTER TABLE public.conversations
    ADD CONSTRAINT conversation_type_check CHECK (conversation_type = ANY (ARRAY[
                'CONVERSATION_CLASS'::text,
                'CONVERSATION_QUESTION'::text,
                'CONVERSATION_COACH'::text,
                'CONVERSATION_LESSON'::text
            ]));


ALTER TABLE conversations
    DROP COLUMN IF EXISTS tutor_id,
    DROP COLUMN IF EXISTS coach_id,
    DROP COLUMN IF EXISTS class_id,
    DROP COLUMN IF EXISTS previous_coach_ids,
    DROP COLUMN IF EXISTS student_id,
    DROP COLUMN IF EXISTS student_question_id,
    DROP COLUMN IF EXISTS guest_id;


ALTER TABLE IF EXISTS conversation_statuses
RENAME TO conversation_members;
---- RUN MIGRATE MANUALLY ONLY

-- insert into conversation_coach (conversation_id,coach_id,created_at,updated_at,previous_coach_ids) (
-- select conversation_id,coach_id,created_at,updated_at,previous_coach_ids from conversations where coach_id is not null)
-- on conflict on constraint conversation_coach_pk do nothing

-- insert into conversation_question (conversation_id,student_question_id, tutor_id, created_at, updated_at) (
-- select conversation_id,student_question_id, tutor_id, created_at, updated_at from conversations where tutor_id is not null)
-- on conflict on constraint conversation_question_pk do nothing

-- insert into conversation_class (conversation_id, class_id, created_at, updated_at) (
-- select conversation_id,class_id, created_at, updated_at from conversations where class_id is not null)
-- on conflict on constraint conversation_class_pk do nothing

