CREATE TABLE public.virtual_classroom_log (
	log_id text NOT NULL,
	lesson_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	is_completed bool NULL,
	attendee_ids _text NOT NULL DEFAULT '{}'::text[],
	total_times_reconnection int4 NULL,
	total_times_updating_room_state int4 NULL,
	total_times_getting_room_state int4 NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__virtual_classroom_log PRIMARY KEY (log_id)
);

CREATE POLICY rls_virtual_classroom_log ON "virtual_classroom_log" USING (permission_check(resource_path, 'virtual_classroom_log'::text)) WITH CHECK (permission_check(resource_path, 'virtual_classroom_log'::text));
CREATE POLICY rls_virtual_classroom_log_restrictive ON "virtual_classroom_log" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'virtual_classroom_log'::text)) WITH CHECK (permission_check(resource_path, 'virtual_classroom_log'::text));

ALTER TABLE "virtual_classroom_log" ENABLE ROW LEVEL security;
ALTER TABLE "virtual_classroom_log" FORCE ROW LEVEL security;