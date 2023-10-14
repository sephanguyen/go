CREATE TABLE IF NOT EXISTS course_access_paths
(
    course_id     TEXT                                                          NOT NULL,
    location_id   TEXT                                                          NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    deleted_at    TIMESTAMP WITH TIME ZONE,
    resource_path TEXT                     DEFAULT autofillresourcepath(),
    CONSTRAINT course_access_paths_pk
        PRIMARY KEY (course_id, location_id)
);
CREATE POLICY rls_course_access_paths ON course_access_paths USING (permission_check(resource_path, 'course_access_paths'))
WITH CHECK (permission_check(resource_path, 'course_access_paths'));

ALTER TABLE "course_access_paths"
    ENABLE ROW LEVEL SECURITY;

ALTER TABLE "course_access_paths"
    FORCE ROW LEVEL SECURITY;