/******
* LIVE_ROOM
********/
CREATE TABLE public.live_room (
	channel_id text NOT NULL,
	channel_name text NOT NULL,
	whiteboard_room_id text NULL,
  ended_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
  resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT live_room_pkey PRIMARY KEY (channel_id),
	CONSTRAINT unique__channel_name UNIQUE (channel_name)
);

CREATE POLICY rls_live_room ON "live_room" 
  USING (permission_check(resource_path, 'live_room')) 
  WITH CHECK (permission_check(resource_path, 'live_room'));

CREATE POLICY rls_live_room_restrictive ON "live_room" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room')) 
  WITH CHECK (permission_check(resource_path, 'live_room'));

ALTER TABLE "live_room" ENABLE ROW LEVEL security;
ALTER TABLE "live_room" FORCE ROW LEVEL security;

/******
* LIVE_ROOM_STATE
********/
CREATE TABLE public.live_room_state (
	live_room_state_id text NOT NULL,
	channel_id text NOT NULL,
	current_material jsonb NULL,
  spotlighted_user text NULL,
	whiteboard_zoom_state jsonb NULL,
	recording jsonb NULL,
	current_polling jsonb NULL,
  created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
  resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT live_room_states_pkey PRIMARY KEY (live_room_state_id),
	CONSTRAINT unique__channel_id UNIQUE (channel_id),
  CONSTRAINT fk__live_room_state__channel_id FOREIGN KEY (channel_id) REFERENCES public.live_room(channel_id)
);

CREATE POLICY rls_live_room_state ON "live_room_state" 
  USING (permission_check(resource_path, 'live_room_state')) 
  WITH CHECK (permission_check(resource_path, 'live_room_state'));

CREATE POLICY rls_live_room_state_restrictive ON "live_room_state" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room_state')) 
  WITH CHECK (permission_check(resource_path, 'live_room_state'));

ALTER TABLE "live_room_state" ENABLE ROW LEVEL security;
ALTER TABLE "live_room_state" FORCE ROW LEVEL security;

/******
* LIVE_ROOM_MEMBER_STATE
********/
CREATE TABLE public.live_room_member_state (
	channel_id text NOT NULL,
	user_id text NOT NULL,
	state_type text NOT NULL,
  string_array_value _text NULL,
	bool_value bool NULL,
  created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT live_room_member_state_pk PRIMARY KEY (channel_id, user_id, state_type),
  CONSTRAINT fk__live_room_member_state__channel_id FOREIGN KEY (channel_id) REFERENCES public.live_room(channel_id)
);

CREATE POLICY rls_live_room_member_state ON "live_room_member_state" 
  USING (permission_check(resource_path, 'live_room_member_state')) 
  WITH CHECK (permission_check(resource_path, 'live_room_member_state'));

CREATE POLICY rls_live_room_member_state_restrictive ON "live_room_member_state" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room_member_state')) 
  WITH CHECK (permission_check(resource_path, 'live_room_member_state'));

ALTER TABLE "live_room_member_state" ENABLE ROW LEVEL security;
ALTER TABLE "live_room_member_state" FORCE ROW LEVEL security;

CREATE INDEX IF NOT EXISTS live_room_member_state__channel_id__idx ON public.live_room_member_state USING btree (channel_id);

/******
* LIVE_ROOM_POLL
********/
CREATE TABLE public.live_room_poll (
	live_room_poll_id text NOT NULL,
	channel_id text NOT NULL,
	"options" jsonb NULL,
	students_answers jsonb NULL,
	stopped_at timestamptz NOT NULL,
	ended_at timestamptz NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT live_room_poll_pk PRIMARY KEY (live_room_poll_id),
	CONSTRAINT fk__live_room_poll__channel_id FOREIGN KEY (channel_id) REFERENCES public.live_room(channel_id)
);

CREATE POLICY rls_live_room_poll ON "live_room_poll" 
  USING (permission_check(resource_path, 'live_room_poll')) 
  WITH CHECK (permission_check(resource_path, 'live_room_poll'));

CREATE POLICY rls_live_room_poll_restrictive ON "live_room_poll" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room_poll')) 
  WITH CHECK (permission_check(resource_path, 'live_room_poll'));

ALTER TABLE "live_room_poll" ENABLE ROW LEVEL security;
ALTER TABLE "live_room_poll" FORCE ROW LEVEL security;

CREATE INDEX IF NOT EXISTS live_room_poll__channel_id__idx ON public.live_room_poll USING btree (channel_id);

/******
* LIVE_ROOM_LOG
********/
CREATE TABLE public.live_room_log (
	live_room_log_id text NOT NULL,
	channel_id text NOT NULL,
	is_completed bool NULL,
	attendee_ids _text NOT NULL DEFAULT '{}'::text[],
	total_times_reconnection int4 NULL,
	total_times_updating_room_state int4 NULL,
	total_times_getting_room_state int4 NULL,
  created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT live_room_log_pk PRIMARY KEY (live_room_log_id),
  CONSTRAINT fk__live_room_log__channel_id FOREIGN KEY (channel_id) REFERENCES public.live_room(channel_id)
);

CREATE POLICY rls_live_room_log ON "live_room_log" 
  USING (permission_check(resource_path, 'live_room_log')) 
  WITH CHECK (permission_check(resource_path, 'live_room_log'));

CREATE POLICY rls_live_room_log_restrictive ON "live_room_log" AS RESTRICTIVE TO PUBLIC 
  USING (permission_check(resource_path, 'live_room_log')) 
  WITH CHECK (permission_check(resource_path, 'live_room_log'));

ALTER TABLE "live_room_log" ENABLE ROW LEVEL security;
ALTER TABLE "live_room_log" FORCE ROW LEVEL security;

CREATE INDEX IF NOT EXISTS live_room_log__channel_id__idx ON public.live_room_log USING btree (channel_id);
