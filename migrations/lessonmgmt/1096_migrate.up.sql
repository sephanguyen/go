CREATE TABLE public.live_lesson_sent_notifications (
	sent_notification_id text NOT NULL,
	lesson_id text NOT NULL,
	sent_at timestamptz NOT NULL,
	sent_at_interval text NOT NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'::text),
	updated_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'::text),
	deleted_at timestamptz NULL,
	CONSTRAINT pk__live_lesson_sent_notifications PRIMARY KEY (sent_notification_id)
);

CREATE INDEX live_lesson_sent_notifications__lesson_id__idx ON public.live_lesson_sent_notifications USING btree (lesson_id);

CREATE POLICY rls_live_lesson_sent_notifications ON "live_lesson_sent_notifications" USING (permission_check(resource_path, 'live_lesson_sent_notifications'::text)) WITH CHECK (permission_check(resource_path, 'live_lesson_sent_notifications'::text));
CREATE POLICY rls_live_lesson_sent_notifications_restrictive ON "live_lesson_sent_notifications" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'live_lesson_sent_notifications'::text)) WITH CHECK (permission_check(resource_path, 'live_lesson_sent_notifications'::text));

ALTER TABLE "live_lesson_sent_notifications" ENABLE ROW LEVEL security;
ALTER TABLE "live_lesson_sent_notifications" FORCE ROW LEVEL security;