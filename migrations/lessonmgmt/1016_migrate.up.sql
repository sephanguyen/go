CREATE TABLE public.lesson_groups (
	lesson_group_id text NOT NULL,
	course_id text NOT NULL,
	media_ids _text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__lesson_groups PRIMARY KEY (lesson_group_id, course_id)
);

CREATE POLICY rls_lesson_groups ON "lesson_groups" USING (permission_check(resource_path, 'lesson_groups')) WITH CHECK (permission_check(resource_path, 'lesson_groups'));
CREATE POLICY rls_lesson_groups_restrictive ON "lesson_groups" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_groups')) with check (permission_check(resource_path, 'lesson_groups'));

ALTER TABLE "lesson_groups" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_groups" FORCE ROW LEVEL security;
