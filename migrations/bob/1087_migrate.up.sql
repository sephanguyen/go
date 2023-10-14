CREATE TABLE IF NOT EXISTS public.lesson_members_states (
    "lesson_id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "state_type" TEXT NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "bool_value" BOOLEAN,
    "resource_path" TEXT,
    CONSTRAINT lesson_id_fk FOREIGN KEY (lesson_id, user_id) REFERENCES public.lesson_members(lesson_id, user_id),
    CONSTRAINT lesson_members_states_pk PRIMARY KEY (lesson_id, user_id, state_type)
);

ALTER TABLE public.lessons
    ADD COLUMN IF NOT EXISTS room_state JSONB;
