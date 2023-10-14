CREATE TABLE IF NOT EXISTS lessons
(
    lesson_id              TEXT                         NOT NULL
        CONSTRAINT lessons_pk
            PRIMARY KEY,
    teacher_id             TEXT,
    course_id              TEXT,
    created_at             TIMESTAMP WITH TIME ZONE     NOT NULL,
    updated_at             TIMESTAMP WITH TIME ZONE     NOT NULL,
    deleted_at             TIMESTAMP WITH TIME ZONE,
    end_at                 TIMESTAMP WITH TIME ZONE,
    control_settings       JSONB,
    lesson_group_id        TEXT,
    room_id                TEXT,
    lesson_type            TEXT,
    status                 TEXT,
    stream_learner_counter INTEGER DEFAULT 0            NOT NULL,
    learner_ids            TEXT[]  DEFAULT '{}'::TEXT[] NOT NULL,
    name                   TEXT,
    start_time             TIMESTAMP WITH TIME ZONE,
    end_time               TIMESTAMP WITH TIME ZONE,
    resource_path          TEXT    DEFAULT autofillresourcepath(),
    room_state             jsonb,
    teaching_model         TEXT,
    class_id               TEXT,
    center_id              TEXT,
    teaching_method        TEXT,
    teaching_medium        TEXT,
    scheduling_status      TEXT    DEFAULT 'LESSON_SCHEDULING_STATUS_PUBLISHED'::TEXT,
    is_locked              BOOLEAN DEFAULT FALSE        NOT NULL,
    scheduler_id           TEXT
);
CREATE POLICY rls_lessons ON lessons USING (permission_check(resource_path, 'lessons'))
WITH CHECK (permission_check(resource_path, 'lessons'));

ALTER TABLE "lessons"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "lessons"
    FORCE ROW LEVEL SECURITY;