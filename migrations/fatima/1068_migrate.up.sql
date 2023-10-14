CREATE TABLE IF NOT EXISTS student_package_class(
    "student_package_id" TEXT NOT NULL,
    "student_id" TEXT NOT NULL,
    "location_id" TEXT NOT NULL DEFAULT '',
    "course_id" TEXT NOT NULL,
    "class_id" TEXT,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" TIMESTAMP WITH TIME ZONE,
    "resource_path" TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT student_package_class_pk PRIMARY KEY (student_package_id, student_id, location_id, course_id, class_id),
    CONSTRAINT student_package_class_fk FOREIGN KEY (student_package_id) REFERENCES "student_packages"(student_package_id)
);

CREATE POLICY rls_student_package_class ON "student_package_class" using (permission_check(resource_path, 'student_package_class')) with check (permission_check(resource_path, 'student_package_class'));

ALTER TABLE "student_package_class" ENABLE ROW LEVEL security;
ALTER TABLE "student_package_class" FORCE ROW LEVEL security;
