CREATE TABLE IF NOT EXISTS public.lesson_recorded_videos (
    recorded_video_id TEXT NOT NULL PRIMARY KEY,
    lesson_id TEXT NOT NULL,
    description TEXT,
    date_time_recorded timestamp with time zone,
    creator TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    media_id TEXT NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT lesson_recorded_videos_lesson_id_fk FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT lesson_recorded_videos_creator_fk FOREIGN KEY (creator) REFERENCES public.users(user_id),
    CONSTRAINT lesson_recorded_videos_media_id_fk FOREIGN KEY (media_id) REFERENCES public.media(media_id),
    CONSTRAINT unique__lesson_id__media_id UNIQUE (lesson_id, media_id)
);

CREATE POLICY rls_lesson_recorded_videos ON "lesson_recorded_videos" using (permission_check(resource_path, 'lesson_recorded_videos')) with check (permission_check(resource_path, 'lesson_recorded_videos'));

ALTER TABLE "lesson_recorded_videos" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_recorded_videos" FORCE ROW LEVEL security;
