-- public.lessons definition
CREATE TABLE IF NOT EXISTS public.lessons (
	lesson_id text NOT NULL,
	teacher_id text NULL,
	course_id text NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	end_at timestamptz NULL,
	control_settings jsonb NULL,
	lesson_group_id text NULL,
	room_id text NULL,
	lesson_type text NULL,
	status text NULL,
	stream_learner_counter int4 NOT NULL DEFAULT 0,
	learner_ids _text NOT NULL DEFAULT '{}'::text[],
	name text NULL,
	start_time timestamptz NULL,
	end_time timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	room_state jsonb NULL,
	teaching_model text NULL,
	class_id text NULL,
	center_id text NULL,
	teaching_method text NULL,
	teaching_medium text NULL,
	scheduling_status text NULL DEFAULT 'LESSON_SCHEDULING_STATUS_PUBLISHED'::text,
	is_locked bool NOT NULL DEFAULT false,
	scheduler_id text NULL,
	classroom_id text NULL,
	CONSTRAINT lessons_pk PRIMARY KEY (lesson_id)
);

CREATE INDEX IF NOT EXISTS lessons__end_time__idx_asc_nulls_last ON public.lessons USING btree (end_time);
CREATE INDEX IF NOT EXISTS lessons__end_time__idx_desc_nulls_first ON public.lessons USING btree (end_time DESC);
CREATE INDEX IF NOT EXISTS lessons__lesson_type__idx ON public.lessons USING btree (lesson_type);
CREATE INDEX IF NOT EXISTS lessons__start_time__idx_asc_nulls_last ON public.lessons USING btree (start_time);
CREATE INDEX IF NOT EXISTS lessons__start_time__idx_desc_nulls_first ON public.lessons USING btree (start_time DESC);

CREATE POLICY rls_lessons ON "lessons" USING (permission_check(resource_path, 'lessons')) WITH CHECK (permission_check(resource_path, 'lessons'));
CREATE POLICY rls_lessons_restrictive ON "lessons" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lessons')) with check (permission_check(resource_path, 'lessons'));

ALTER TABLE "lessons" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "lessons" FORCE ROW LEVEL SECURITY;