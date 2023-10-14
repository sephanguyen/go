CREATE TABLE public.courses_academic_years (
	course_id text NOT NULL,
	academic_year_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__courses_academic_years PRIMARY KEY (course_id, academic_year_id)
);

CREATE POLICY rls_courses_academic_years ON "courses_academic_years" USING (permission_check(resource_path, 'courses_academic_years'::text)) WITH CHECK (permission_check(resource_path, 'courses_academic_years'::text));
CREATE POLICY rls_courses_academic_years_restrictive ON "courses_academic_years" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'courses_academic_years'::text)) WITH CHECK (permission_check(resource_path, 'courses_academic_years'::text));

ALTER TABLE "courses_academic_years" ENABLE ROW LEVEL security;
ALTER TABLE "courses_academic_years" FORCE ROW LEVEL security;