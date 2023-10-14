CREATE TABLE public.student_parents (
	student_id text NOT NULL,
	parent_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	relationship text NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT student_parents_pk PRIMARY KEY (student_id, parent_id)
);

CREATE INDEX student_parents__parent_id__idx ON public.student_parents USING btree (parent_id);
CREATE INDEX student_parents__student_id_idx ON public.student_parents USING btree (student_id);

CREATE POLICY rls_student_parents ON "student_parents" USING (permission_check(resource_path, 'student_parents'::text)) WITH CHECK (permission_check(resource_path, 'student_parents'::text));
CREATE POLICY rls_student_parents_restrictive ON "student_parents" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'student_parents'::text)) WITH CHECK (permission_check(resource_path, 'student_parents'::text));

ALTER TABLE "student_parents" ENABLE ROW LEVEL security;
ALTER TABLE "student_parents" FORCE ROW LEVEL security;