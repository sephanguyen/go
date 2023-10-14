-- public.lessons_teachers definition
CREATE TABLE IF NOT EXISTS public.lessons_teachers (
	lesson_id text NOT NULL,
	teacher_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	teacher_name text NULL,
	CONSTRAINT lessons_teachers_pk PRIMARY KEY (lesson_id, teacher_id)
);

-- public.lessons_teachers foreign keys
ALTER TABLE public.lessons_teachers ADD CONSTRAINT lessons_fk FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id);

CREATE POLICY rls_lessons_teachers ON lessons_teachers USING (permission_check(resource_path, 'lessons_teachers')) WITH CHECK (permission_check(resource_path, 'lessons_teachers'));
CREATE POLICY rls_lessons_teachers_restrictive ON "lessons_teachers" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lessons_teachers')) WITH CHECK (permission_check(resource_path, 'lessons_teachers'));

ALTER TABLE "lessons_teachers" ENABLE ROW LEVEL SECURITY; 
ALTER TABLE "lessons_teachers" FORCE ROW LEVEL SECURITY;