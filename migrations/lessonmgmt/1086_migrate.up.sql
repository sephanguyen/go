CREATE TABLE public.activity_logs (
	activity_log_id text NOT NULL,
	user_id text NOT NULL,
	action_type text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	payload jsonb NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT activity_logs_pk PRIMARY KEY (activity_log_id)
);

CREATE INDEX activity_logs_payload ON public.activity_logs USING gin (payload);

CREATE POLICY rls_activity_logs ON "activity_logs" USING (permission_check(resource_path, 'activity_logs'::text)) WITH CHECK (permission_check(resource_path, 'activity_logs'::text));
CREATE POLICY rls_activity_logs_restrictive ON "activity_logs" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'activity_logs'::text)) WITH CHECK (permission_check(resource_path, 'activity_logs'::text));

ALTER TABLE "activity_logs" ENABLE ROW LEVEL security;
ALTER TABLE "activity_logs" FORCE ROW LEVEL security;