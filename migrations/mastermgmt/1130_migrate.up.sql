CREATE TABLE IF NOT EXISTS public.course_academic_year (
	course_id text NOT NULL,
    academic_year_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT course_academic_year_pkey PRIMARY KEY (academic_year_id,course_id)
);

CREATE POLICY rls_course_academic_year ON "course_academic_year" using (permission_check(resource_path, 'course_academic_year')) with check (permission_check(resource_path, 'course_academic_year'));
CREATE POLICY rls_course_academic_year_restrictive ON "course_academic_year" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'course_academic_year')) with check (permission_check(resource_path, 'course_academic_year'));

ALTER TABLE "course_academic_year" ENABLE ROW LEVEL security;
ALTER TABLE "course_academic_year" FORCE ROW LEVEL security;

DO
$$
BEGIN
  IF NOT is_table_in_publication('debezium_publication', 'course_academic_year') THEN
    ALTER PUBLICATION debezium_publication ADD TABLE public.course_academic_year;
  END IF;
END;
$$;
