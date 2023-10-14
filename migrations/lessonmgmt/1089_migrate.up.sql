CREATE TABLE public.lesson_polls (
	poll_id text NOT NULL,
	lesson_id text NOT NULL,
	"options" jsonb NULL,
	students_answers jsonb NULL,
	stopped_at timestamptz NOT NULL,
	ended_at timestamptz NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT lesson_polls_pk PRIMARY KEY (poll_id)
);

CREATE POLICY rls_lesson_polls ON "lesson_polls" USING (permission_check(resource_path, 'lesson_polls'::text)) WITH CHECK (permission_check(resource_path, 'lesson_polls'::text));
CREATE POLICY rls_lesson_polls_restrictive ON "lesson_polls" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'lesson_polls'::text)) WITH CHECK (permission_check(resource_path, 'lesson_polls'::text));

ALTER TABLE "lesson_polls" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_polls" FORCE ROW LEVEL security;