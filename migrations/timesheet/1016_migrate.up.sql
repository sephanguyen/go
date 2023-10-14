CREATE TABLE IF NOT EXISTS timesheet_lesson_hours
(
    timesheet_id  TEXT  NOT NULL,
    lesson_id     TEXT  NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at    TIMESTAMP WITH TIME ZONE,
    resource_path TEXT                     DEFAULT autofillresourcepath(),
    CONSTRAINT timesheet_lesson_hours_pk
        PRIMARY KEY (timesheet_id, lesson_id)
);
CREATE POLICY rls_timesheet_lesson_hours ON timesheet_lesson_hours USING (permission_check(resource_path, 'timesheet_lesson_hours'))
WITH CHECK (permission_check(resource_path, 'timesheet_lesson_hours'));

ALTER TABLE "timesheet_lesson_hours"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet_lesson_hours"
    FORCE ROW LEVEL SECURITY;