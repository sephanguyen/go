CREATE TABLE IF NOT EXISTS public.live_lesson_sent_notifications(
  sent_notification_id text not null,
  lesson_id text not null,  
  sent_at timestamp with time zone NOT NULL,
  sent_at_interval text not null,
  resource_path text default autofillresourcepath(),
  created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
  updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
  deleted_at timestamp with time zone,  
  CONSTRAINT pk__live_lesson_sent_notifications PRIMARY KEY(sent_notification_id));


CREATE POLICY rls_live_lesson_sent_notifications ON "live_lesson_sent_notifications"
    USING (permission_check (resource_path, 'live_lesson_sent_notifications'))
    WITH CHECK (permission_check (resource_path, 'live_lesson_sent_notifications'));

CREATE POLICY rls_live_lesson_sent_notifications_restrictive ON "live_lesson_sent_notifications" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'live_lesson_sent_notifications'))
WITH CHECK (permission_check(resource_path, 'live_lesson_sent_notifications'));

ALTER TABLE "live_lesson_sent_notifications" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "live_lesson_sent_notifications" FORCE ROW LEVEL SECURITY;