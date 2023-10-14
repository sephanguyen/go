CREATE TABLE IF NOT EXISTS student_parents
(
	student_id TEXT,
	parent_id TEXT,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path text DEFAULT autofillresourcepath(),
	CONSTRAINT student_parents_pk PRIMARY KEY(student_id, parent_id)
);

ALTER TABLE public.student_parents ADD COLUMN IF NOT EXISTS relationship text NOT NULL ;

CREATE POLICY rls_student_parents ON "student_parents" using (permission_check(resource_path, 'student_parents')) with check (permission_check(resource_path, 'student_parents'));

ALTER TABLE "student_parents" ENABLE ROW LEVEL security;
ALTER TABLE "student_parents" FORCE ROW LEVEL security;