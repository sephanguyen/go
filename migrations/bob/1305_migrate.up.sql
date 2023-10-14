ALTER TABLE only public.users_info_notifications ADD COLUMN IF NOT EXISTS grade_id TEXT;
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

CREATE TABLE IF NOT EXISTS public.notification_grade_mapping (
    id                      TEXT NOT NULL,
    grade_id                TEXT NOT NULL,
    map_value               SMALLINT NOT NULL,
    resource_path           TEXT NOT NULL DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,

    CONSTRAINT pk__notification_grade_mapping PRIMARY KEY (id)
);

CREATE POLICY rls_notification_grade_mapping ON "notification_grade_mapping"
USING (permission_check(resource_path, 'notification_grade_mapping'))
WITH CHECK (permission_check(resource_path, 'notification_grade_mapping'));

CREATE POLICY rls_notification_grade_mapping_restrictive ON "notification_grade_mapping" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'notification_grade_mapping'))
with check (permission_check(resource_path, 'notification_grade_mapping'));

ALTER TABLE IF EXISTS public.notification_grade_mapping ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.notification_grade_mapping FORCE ROW LEVEL security;