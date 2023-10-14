CREATE TABLE public.lesson_recorded_videos (
	recorded_video_id text NOT NULL,
	lesson_id text NOT NULL,
	description text NULL,
	date_time_recorded timestamptz NULL,
	creator text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	media_id text NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT lesson_recorded_videos_pkey PRIMARY KEY (recorded_video_id),
	CONSTRAINT unique__lesson_id__media_id UNIQUE (lesson_id, media_id)
);

CREATE POLICY rls_lesson_recorded_videos ON "lesson_recorded_videos" USING (permission_check(resource_path, 'lesson_recorded_videos'::text)) WITH CHECK (permission_check(resource_path, 'lesson_recorded_videos'::text));
CREATE POLICY rls_lesson_recorded_videos_restrictive ON "lesson_recorded_videos" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'lesson_recorded_videos'::text)) WITH CHECK (permission_check(resource_path, 'lesson_recorded_videos'::text));

ALTER TABLE "lesson_recorded_videos" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_recorded_videos" FORCE ROW LEVEL security;