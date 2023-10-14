ALTER TABLE ONLY public.lessons
    ADD COLUMN IF NOT EXISTS "preparation_time" INTEGER,
    ADD COLUMN IF NOT EXISTS "break_time" INTEGER;
    
-- create new course_time table
CREATE TABLE IF NOT EXISTS public.course_teaching_time (
	course_id text NOT NULL,
	preparation_time INTEGER,
	break_time INTEGER,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT course_id_pk PRIMARY KEY (course_id),
    CONSTRAINT course_id_fk FOREIGN KEY (course_id) REFERENCES public.courses (course_id)
);

CREATE POLICY rls_course_teaching_time ON "course_teaching_time" USING (permission_check(resource_path, 'course_teaching_time')) WITH CHECK (permission_check(resource_path, 'course_teaching_time'));
CREATE POLICY rls_course_teaching_time_restrictive ON "course_teaching_time" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'course_teaching_time')) with check (permission_check(resource_path, 'course_teaching_time'));

ALTER TABLE "course_teaching_time" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "course_teaching_time" FORCE ROW LEVEL SECURITY;

ALTER PUBLICATION debezium_publication ADD TABLE public.course_teaching_time;
