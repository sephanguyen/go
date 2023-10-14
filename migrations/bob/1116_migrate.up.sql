CREATE TABLE IF NOT EXISTS public.lesson_polls (
    poll_id text UNIQUE NOT NULL,
    lesson_id TEXT NOT NULL,
    options JSONB,
    students_answers JSONB,
    stopped_at timestamp with time zone NOT NULL,
    ended_at TIMESTAMP,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP,
    CONSTRAINT lessons_fk FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT lesson_polls_pk PRIMARY KEY (poll_id)
);


ALTER TABLE lesson_members_states
    ADD COLUMN IF NOT EXISTS string_array_value TEXT[] DEFAULT NULL;
