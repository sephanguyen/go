CREATE TABLE IF NOT EXISTS student_package_access_path(
    "student_package_id" TEXT NOT NULL,
    "course_id" TEXT NOT NULL,
    "student_id" TEXT NOT NULL,
    "location_id" TEXT NOT NULL DEFAULT '',
    "access_path" TEXT,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" TIMESTAMP WITH TIME ZONE,
    "resource_path" TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT student_package_access_path_pk PRIMARY KEY (student_package_id, course_id, student_id, location_id),
    CONSTRAINT student_package_access_path_student_packages_fk FOREIGN KEY (student_package_id) REFERENCES "student_packages"(student_package_id)
);

CREATE POLICY rls_student_package_access_path ON "student_package_access_path" using (permission_check(resource_path, 'student_package_access_path')) with check (permission_check(resource_path, 'student_package_access_path'));

ALTER TABLE "student_package_access_path" ENABLE ROW LEVEL security;
ALTER TABLE "student_package_access_path" FORCE ROW LEVEL security;