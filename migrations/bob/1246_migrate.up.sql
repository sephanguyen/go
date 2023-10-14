-- public.student_course_slot definition
CREATE TABLE IF NOT EXISTS public.student_course_slot (
	student_course_slot_id text NOT NULL,
	student_id text NOT NULL,
	location_id text NOT NULL,
	course_id text NOT NULL,
	student_package_id text NOT NULL,
	student_start_date date NOT NULL,
    student_end_date date NOT NULL,
    course_slot int4 NULL,
    course_slot_per_week int4 NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    CONSTRAINT student_course_slot_pk PRIMARY KEY (student_course_slot_id),
    CONSTRAINT student_course_slot_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id),
    CONSTRAINT student_course_slot_courses_fk FOREIGN KEY (course_id) REFERENCES "courses"(course_id),
    CONSTRAINT student_course_slot_locations_fk FOREIGN KEY (location_id) REFERENCES "locations"(location_id)
);

CREATE POLICY rls_student_course_slot ON "student_course_slot" using (permission_check(resource_path, 'student_course_slot')) with check (permission_check(resource_path, 'student_course_slot'));
ALTER TABLE "student_course_slot" ENABLE ROW LEVEL security;
ALTER TABLE "student_course_slot" FORCE ROW LEVEL security;
