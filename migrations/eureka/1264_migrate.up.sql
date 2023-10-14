CREATE TABLE IF NOT EXISTS courses
(
    course_id            TEXT                                    NOT NULL
        CONSTRAINT courses_pk
            PRIMARY KEY,
    name                 TEXT                                    NOT NULL,
    country              TEXT,
    subject              TEXT,
    grade                SMALLINT,
    display_order        SMALLINT DEFAULT 0,
    updated_at           TIMESTAMP WITH TIME ZONE                NOT NULL,
    created_at           TIMESTAMP WITH TIME ZONE                NOT NULL,
    school_id            INTEGER  DEFAULT '-2147483648'::INTEGER NOT NULL,
    deleted_at           TIMESTAMP WITH TIME ZONE,
    course_type          TEXT,
    start_date           TIMESTAMP WITH TIME ZONE,
    end_date             TIMESTAMP WITH TIME ZONE,
    teacher_ids          TEXT[],
    preset_study_plan_id TEXT,
    icon                 TEXT,
    status               TEXT     DEFAULT 'COURSE_STATUS_NONE'::TEXT,
    resource_path        TEXT     DEFAULT autofillresourcepath(),
    teaching_method      TEXT,
    course_type_id       TEXT,
    remarks              TEXT,
    is_archived          boolean DEFAULT false,
    course_partner_id    TEXT
);

CREATE POLICY rls_courses ON "courses" USING (
    permission_check(resource_path, 'courses'))
WITH CHECK (
    permission_check(resource_path, 'courses'));

CREATE POLICY rls_courses_restrictive ON "courses" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'courses'))
WITH CHECK (
    permission_check(resource_path, 'courses'));

ALTER TABLE "courses" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "courses" FORCE ROW LEVEL SECURITY;
