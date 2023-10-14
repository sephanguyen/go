CREATE TABLE public.lesson_classrooms (
	lesson_id text NOT NULL,
	classroom_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__lesson_classrooms PRIMARY KEY (lesson_id,classroom_id)
);


ALTER TABLE public.lesson_classrooms ENABLE ROW LEVEL security;
ALTER TABLE public.lesson_classrooms FORCE ROW LEVEL security;

CREATE POLICY rls_lesson_classrooms ON "lesson_classrooms" USING (permission_check(resource_path, 'lesson_classrooms')) WITH CHECK (permission_check(resource_path, 'lesson_classrooms'));
CREATE POLICY rls_lesson_classrooms_restrictive ON "lesson_classrooms" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_classrooms')) with check (permission_check(resource_path, 'lesson_classrooms'));

