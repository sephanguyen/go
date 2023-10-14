-- table to backup notification target group data
CREATE TABLE IF NOT EXISTS public.notification_target_group (
	notification_id     TEXT NOT NULL,
	target_groups       JSONB NULL,
	resource_path       TEXT NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__notification_target_group PRIMARY KEY (notification_id),

	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone
);

CREATE POLICY rls_notification_target_group ON "notification_target_group"
USING (permission_check(resource_path, 'notification_target_group'))
WITH CHECK (permission_check(resource_path, 'notification_target_group'));

CREATE POLICY rls_notification_target_group_restrictive ON "notification_target_group" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'notification_target_group'))
with check (permission_check(resource_path, 'notification_target_group'));

ALTER TABLE IF EXISTS public.notification_target_group ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.notification_target_group FORCE ROW LEVEL security;