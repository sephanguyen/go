CREATE TABLE IF NOT EXISTS public.lesson_room_states (
    lesson_room_state_id TEXT NOT NULL PRIMARY KEY,
    lesson_id TEXT NOT NULL,
    current_material jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    agora_room_id TEXT,
    streaming_attendees TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
    ended_at timestamp with time zone,
    spotlighted_user TEXT,
    whiteboard_zoom_state jsonb,
    recording jsonb,
    current_polling jsonb,
    CONSTRAINT lesson_id_fk FOREIGN KEY (lesson_id) REFERENCES lessons(lesson_id),
    CONSTRAINT unique__lesson_id UNIQUE (lesson_id)
);

DROP POLICY IF EXISTS rls_lesson_room_states ON public.lesson_room_states;
CREATE POLICY rls_lesson_room_states ON "lesson_room_states" USING (permission_check(resource_path, 'lesson_room_states'::text)) WITH CHECK (permission_check(resource_path, 'lesson_room_states'::text));

DROP POLICY IF EXISTS rls_lesson_room_states_restrictive ON public.lesson_room_states;
CREATE POLICY rls_lesson_room_states_restrictive ON "lesson_room_states" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'lesson_room_states'::text)) WITH CHECK (permission_check(resource_path, 'lesson_room_states'::text));

ALTER TABLE "lesson_room_states" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_room_states" FORCE ROW LEVEL security;
