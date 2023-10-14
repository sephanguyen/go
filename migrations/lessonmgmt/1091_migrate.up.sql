CREATE TABLE public.academic_years (
	academic_year_id text NOT NULL,
	school_id int4 NOT NULL,
	"name" text NOT NULL,
	start_year_date timestamptz NOT NULL,
	end_year_date timestamptz NOT NULL,
	status text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__academic_years PRIMARY KEY (academic_year_id)
);

CREATE POLICY rls_academic_years ON "academic_years" USING (permission_check(resource_path, 'academic_years'::text)) WITH CHECK (permission_check(resource_path, 'academic_years'::text));
CREATE POLICY rls_academic_years_restrictive ON "academic_years" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'academic_years'::text)) WITH CHECK (permission_check(resource_path, 'academic_years'::text));

ALTER TABLE "academic_years" ENABLE ROW LEVEL security;
ALTER TABLE "academic_years" FORCE ROW LEVEL security;