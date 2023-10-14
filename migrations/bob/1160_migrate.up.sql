CREATE TABLE IF NOT EXISTS "course_access_paths" (
    "course_id" text NOT NULL,
    "location_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),
    CONSTRAINT course_access_paths_courses_fk FOREIGN KEY (course_id) REFERENCES "courses"(course_id),
    CONSTRAINT course_access_paths_locations_fk FOREIGN KEY (location_id) REFERENCES "locations"(location_id),
    CONSTRAINT course_access_paths_pk PRIMARY KEY (course_id, location_id)
);

CREATE POLICY rls_course_access_paths ON "course_access_paths" using (permission_check(resource_path, 'course_access_paths')) with check (permission_check(resource_path, 'course_access_paths'));

ALTER TABLE "course_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "course_access_paths" FORCE ROW LEVEL security;