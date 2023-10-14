CREATE TABLE IF NOT EXISTS public.virtual_classroom_log (
    log_id TEXT NOT NULL,
    lesson_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    is_completed BOOLEAN,
    attendee_ids TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
    total_times_reconnection INTEGER,
    total_times_updating_room_state INTEGER,
    total_times_getting_room_state INTEGER,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__virtual_classroom_log PRIMARY KEY (log_id),
    CONSTRAINT fk__virtual_classroom_log__lesson_id FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id)
);

CREATE POLICY rls_virtual_classroom_log ON "virtual_classroom_log" using (permission_check(resource_path, 'virtual_classroom_log')) with check (permission_check(resource_path, 'virtual_classroom_log'));

ALTER TABLE "virtual_classroom_log" ENABLE ROW LEVEL security;
ALTER TABLE "virtual_classroom_log" FORCE ROW LEVEL security;
