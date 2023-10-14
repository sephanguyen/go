-- public.lesson_members definition
CREATE TABLE IF NOT EXISTS public.lesson_members (
	lesson_id text NOT NULL,
	user_id text NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	attendance_status text NULL,
	attendance_remark text NULL,
	course_id text NULL,
	attendance_notice text NULL,
	attendance_reason text NULL,
	attendance_note text NULL,
	user_first_name text NULL,
	user_last_name text NULL,
	CONSTRAINT pk__lesson_members PRIMARY KEY (lesson_id, user_id)
);
-- public.lesson_members foreign keys
ALTER TABLE public.lesson_members ADD CONSTRAINT fk__lesson_members__lesson_id FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id);

CREATE POLICY rls_lesson_members ON lesson_members USING (permission_check(resource_path, 'lesson_members')) 
WITH CHECK (permission_check(resource_path, 'lesson_members'));
CREATE POLICY rls_lesson_members_restrictive ON "lesson_members" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_members')) with check (permission_check(resource_path, 'lesson_members'));

ALTER TABLE "lesson_members" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "lesson_members" FORCE ROW LEVEL SECURITY;