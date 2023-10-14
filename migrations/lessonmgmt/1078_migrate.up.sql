/******
* LIVE_LESSON_CONVERSATION
********/
CREATE TABLE IF NOT EXISTS public.live_lesson_conversation (
    lesson_conversation_id text NOT NULL,
    conversation_id text NOT NULL UNIQUE,
    lesson_id text NOT NULL,
    participant_list _text NOT NULL DEFAULT '{}'::text[],
    created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT live_lesson_conversation_pkey PRIMARY KEY (lesson_conversation_id)
);

CREATE INDEX IF NOT EXISTS live_lesson_conversation__lesson_id__idx ON public.live_lesson_conversation USING btree (lesson_id);


DROP POLICY IF EXISTS rls_live_lesson_conversation ON public.live_lesson_conversation;

CREATE POLICY rls_live_lesson_conversation ON "live_lesson_conversation"
  USING (permission_check(resource_path, 'live_lesson_conversation'))
  WITH CHECK (permission_check(resource_path, 'live_lesson_conversation'));


DROP POLICY IF EXISTS rls_live_lesson_conversation_restrictive ON public.live_lesson_conversation;
CREATE POLICY rls_live_lesson_conversation_restrictive ON "live_lesson_conversation" AS RESTRICTIVE TO PUBLIC
  USING (permission_check(resource_path, 'live_lesson_conversation'))
  WITH CHECK (permission_check(resource_path, 'live_lesson_conversation'));

ALTER TABLE "live_lesson_conversation" ENABLE ROW LEVEL security;
ALTER TABLE "live_lesson_conversation" FORCE ROW LEVEL security;
----------------------------------------------
CREATE TABLE IF NOT EXISTS public.student_course (
    student_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    student_package_id TEXT NOT NULL,
    package_type TEXT NULL,
    course_slot int4 NULL,
    course_slot_per_week int4 NULL,
    weight int4 NULL,
    student_start_date timestamp with time zone NOT NULL,
    student_end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath()
    );

ALTER TABLE public.student_course DROP CONSTRAINT IF EXISTS student_course_pk;
ALTER TABLE public.student_course ADD CONSTRAINT student_course_pk PRIMARY KEY (student_id, course_id, location_id, student_package_id);

CREATE INDEX IF NOT EXISTS student_course__package_type__idx ON public.student_course USING btree (package_type);

DROP POLICY IF EXISTS rls_student_course ON student_course;
CREATE POLICY rls_student_course ON "student_course" USING (permission_check(resource_path, 'student_course'::text)) WITH CHECK (permission_check(resource_path, 'student_course'::text));


DROP POLICY IF EXISTS rls_student_course_restrictive ON student_course;
CREATE POLICY rls_student_course_restrictive ON "student_course" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'student_course'::text)) WITH CHECK (permission_check(resource_path, 'student_course'::text));

ALTER TABLE "student_course" ENABLE ROW LEVEL security;
ALTER TABLE "student_course" FORCE ROW LEVEL security;
