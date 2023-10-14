/******
* LIVE_ROOM_STATE
********/
ALTER TABLE IF EXISTS public.live_room_state
    ADD COLUMN IF NOT EXISTS stream_learner_counter int4 NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS streaming_learners _text NOT NULL DEFAULT '{}'::text[];

/******
* LIVE_ROOM_ACTIVITY_LOGS
********/
CREATE TABLE public.live_room_activity_logs (
    activity_log_id text NOT NULL,
    channel_id text NOT NULL,
    user_id text NOT NULL,
    action_type text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT live_room_activity_logs_pkey PRIMARY KEY (activity_log_id),
    CONSTRAINT fk__live_room_activity_logs__channel_id FOREIGN KEY (channel_id) REFERENCES public.live_room(channel_id)
);

CREATE POLICY rls_live_room_activity_logs ON "live_room_activity_logs" 
  USING (permission_check(resource_path, 'live_room_activity_logs')) 
  WITH CHECK (permission_check(resource_path, 'live_room_activity_logs'));

CREATE POLICY rls_live_room_activity_logs_restrictive ON "live_room_activity_logs" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room_activity_logs')) 
  WITH CHECK (permission_check(resource_path, 'live_room_activity_logs'));

ALTER TABLE "live_room_activity_logs" ENABLE ROW LEVEL security;
ALTER TABLE "live_room_activity_logs" FORCE ROW LEVEL security;