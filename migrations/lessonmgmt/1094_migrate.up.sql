CREATE TABLE public.courses_classes (
	course_id text NOT NULL,
	class_id int4 NOT NULL,
	status text NOT NULL DEFAULT 'COURSE_CLASS_STATUS_ACTIVE'::text,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT courses_classes_pk PRIMARY KEY (course_id, class_id)
);

CREATE INDEX courses_classes_class_id_idx ON public.courses_classes USING btree (class_id);

CREATE POLICY rls_courses_classes ON "courses_classes" USING (permission_check(resource_path, 'courses_classes'::text)) WITH CHECK (permission_check(resource_path, 'courses_classes'::text));
CREATE POLICY rls_courses_classes_restrictive ON "courses_classes" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'courses_classes'::text)) WITH CHECK (permission_check(resource_path, 'courses_classes'::text));

ALTER TABLE "courses_classes" ENABLE ROW LEVEL security;
ALTER TABLE "courses_classes" FORCE ROW LEVEL security;