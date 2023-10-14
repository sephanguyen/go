ALTER TABLE public.lessons DROP CONSTRAINT IF EXISTS lesson_classroom_fk;

CREATE TABLE IF NOT EXISTS public.lesson_classrooms (
    lesson_id TEXT NOT NULL,
    classroom_id TEXT NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT pk__lesson_classrooms PRIMARY KEY (lesson_id,classroom_id),
    CONSTRAINT fk__lesson_classrooms__lesson_id FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT fk__lesson_classrooms__classroom_id FOREIGN KEY (classroom_id) REFERENCES public.classroom(classroom_id)
);
CREATE INDEX lesson_classrooms__classroom_id__idx on public.lesson_classrooms (classroom_id);

CREATE POLICY rls_lesson_classrooms ON "lesson_classrooms" using (permission_check(resource_path, 'lesson_classrooms')) with check (permission_check(resource_path, 'lesson_classrooms'));
CREATE POLICY rls_lesson_classrooms_restrictive ON "lesson_classrooms" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_classrooms')) with check (permission_check(resource_path, 'lesson_classrooms'));
ALTER TABLE "lesson_classrooms" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_classrooms" FORCE ROW LEVEL security;

