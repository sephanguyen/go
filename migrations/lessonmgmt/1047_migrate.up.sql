/******
* LIVE_ROOM_RECORDED_VIDEOS
********/
CREATE TABLE public.live_room_recorded_videos (
    recorded_video_id TEXT NOT NULL PRIMARY KEY,
	channel_id text NOT NULL,
    media_id TEXT NOT NULL,
	description TEXT,
    date_time_recorded timestamp with time zone,
    creator TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT unique__channel_id_media_id UNIQUE (channel_id, media_id),
    CONSTRAINT fk__live_room_recorded_videos__channel_id FOREIGN KEY (channel_id) REFERENCES public.live_room(channel_id)
);

CREATE POLICY rls_live_room_recorded_videos ON "live_room_recorded_videos" 
  USING (permission_check(resource_path, 'live_room_recorded_videos')) 
  WITH CHECK (permission_check(resource_path, 'live_room_recorded_videos'));

CREATE POLICY rls_live_room_recorded_videos_restrictive ON "live_room_recorded_videos" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room_recorded_videos')) 
  WITH CHECK (permission_check(resource_path, 'live_room_recorded_videos'));

ALTER TABLE "live_room_recorded_videos" ENABLE ROW LEVEL security;
ALTER TABLE "live_room_recorded_videos" FORCE ROW LEVEL security;
