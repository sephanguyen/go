CREATE TABLE IF NOT EXISTS public.lesson_room_states (
    lesson_room_state_id TEXT NOT NULL PRIMARY KEY,
    lesson_id text NOT NULL,
    current_material jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT lesson_id_fk FOREIGN KEY (lesson_id) REFERENCES lessons(lesson_id),
    CONSTRAINT unique__lesson_id UNIQUE (lesson_id)
);

CREATE POLICY rls_lesson_room_states ON "lesson_room_states" using (permission_check(resource_path, 'lesson_room_states')) with check (permission_check(resource_path, 'lesson_room_states'));

ALTER TABLE "lesson_room_states" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_room_states" FORCE ROW LEVEL security;