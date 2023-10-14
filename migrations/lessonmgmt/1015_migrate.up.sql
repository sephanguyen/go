CREATE TABLE public.reallocation (
	student_id text NOT NULL,
	original_lesson_id text NOT NULL,
	new_lesson_id text NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	course_id text NOT NULL,
	CONSTRAINT reallocation__pk PRIMARY KEY (student_id, original_lesson_id)
);

CREATE POLICY rls_reallocation ON "reallocation" USING (permission_check(resource_path, 'reallocation')) WITH CHECK (permission_check(resource_path, 'reallocation'));
CREATE POLICY rls_reallocation_restrictive ON "reallocation" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'reallocation')) with check (permission_check(resource_path, 'reallocation'));

ALTER TABLE "reallocation" ENABLE ROW LEVEL security;
ALTER TABLE "reallocation" FORCE ROW LEVEL security;
