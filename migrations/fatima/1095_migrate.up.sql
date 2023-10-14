CREATE TABLE public.student_course (
    student_id text NOT NULL,
    course_id text NULL,
    location_id text NOT NULL,
    student_package_id text NOT NULL,
    student_start_date timestamp with time zone NOT NULL,
    student_end_date timestamp with time zone NOT NULL,
    course_slot integer,
    course_slot_per_week integer,
    weight integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT student_course_pk PRIMARY KEY (student_id, course_id, location_id, student_package_id),
    CONSTRAINT student_course_student_id_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id),
    CONSTRAINT student_course_course_id_fk FOREIGN KEY (course_id) REFERENCES "courses"(course_id),
    CONSTRAINT student_course_location_id_fk FOREIGN KEY (location_id) REFERENCES "locations"(location_id),
    CONSTRAINT student_course_student_package_id_fk FOREIGN KEY (student_package_id) REFERENCES "student_package_by_order"(student_package_id)
);

CREATE POLICY rls_student_course ON "student_course" USING (permission_check(resource_path, 'student_course')) WITH CHECK (permission_check(resource_path, 'student_course'));

CREATE POLICY rls_student_course_restrictive ON "student_course"  AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'student_course'))
WITH CHECK (permission_check(resource_path, 'student_course'));

ALTER TABLE "student_course" ENABLE ROW LEVEL security;
ALTER TABLE "student_course" FORCE ROW LEVEL security;