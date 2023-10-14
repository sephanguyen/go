-- public.students definition

-- Drop table

-- DROP TABLE public.students;

CREATE TABLE IF NOT EXISTS public.students (
	student_id text NOT NULL,
	current_grade int2 NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	student_external_id text NULL,
	grade_id text NULL,
	CONSTRAINT students_pk PRIMARY KEY (student_id)
);

CREATE POLICY rls_students ON "students" USING (permission_check(resource_path, 'students')) WITH CHECK (permission_check(resource_path, 'students'));
ALTER TABLE "students"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "students"
    FORCE ROW LEVEL SECURITY;

-- public.grade definition

-- Drop table

-- DROP TABLE public.grade;

CREATE TABLE IF NOT EXISTS public.grade (
	"name" text NOT NULL,
	is_archived bool NOT NULL DEFAULT false,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	grade_id text NOT NULL,
	deleted_at timestamptz NULL,
	"sequence" int4 NULL,
	CONSTRAINT grade_pk PRIMARY KEY (grade_id)
);

CREATE POLICY rls_grade ON "grade" USING (permission_check(resource_path, 'grade')) WITH CHECK (permission_check(resource_path, 'grade'));
ALTER TABLE IF EXISTS "grade"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS "grade"
    FORCE ROW LEVEL SECURITY;
