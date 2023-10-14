-- notification_course_filter
CREATE TABLE IF NOT EXISTS notification_course_filter(
    notification_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone DEFAULT timezone('utc'::text, now()),
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk_notification_course_filter PRIMARY KEY (notification_id, course_id),
    CONSTRAINT fk_notification_course_filter__notification_id FOREIGN KEY (notification_id) REFERENCES public.info_notifications(notification_id)
);

CREATE POLICY rls_notification_course_filter ON "notification_course_filter"
USING (permission_check(resource_path, 'notification_course_filter'))
WITH CHECK (permission_check(resource_path, 'notification_course_filter'));

CREATE POLICY rls_notification_course_filter_restrictive ON "notification_course_filter" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'notification_course_filter'))
with check (permission_check(resource_path, 'notification_course_filter'));

ALTER TABLE "notification_course_filter" ENABLE ROW LEVEL security;
ALTER TABLE "notification_course_filter" FORCE ROW LEVEL security;


-- notification_location_filter
CREATE TABLE IF NOT EXISTS notification_location_filter(
    notification_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone DEFAULT timezone('utc'::text, now()),
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk_notification_location_filter PRIMARY KEY (notification_id, location_id),
    CONSTRAINT fk_notification_location_filter__notification_id FOREIGN KEY (notification_id) REFERENCES public.info_notifications(notification_id)
);

CREATE POLICY rls_notification_location_filter ON "notification_location_filter"
USING (permission_check(resource_path, 'notification_location_filter'))
WITH CHECK (permission_check(resource_path, 'notification_location_filter'));

CREATE POLICY rls_notification_location_filter_restrictive ON "notification_location_filter" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'notification_location_filter'))
with check (permission_check(resource_path, 'notification_location_filter'));

ALTER TABLE "notification_location_filter" ENABLE ROW LEVEL security;
ALTER TABLE "notification_location_filter" FORCE ROW LEVEL security;

-- notification_class_filter
CREATE TABLE IF NOT EXISTS notification_class_filter(
    notification_id TEXT NOT NULL,
    class_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone DEFAULT timezone('utc'::text, now()),
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk_notification_class_filter PRIMARY KEY (notification_id, class_id),
    CONSTRAINT fk_notification_class_filter__notification_id FOREIGN KEY (notification_id) REFERENCES public.info_notifications(notification_id)
);

CREATE POLICY rls_notification_class_filter ON "notification_class_filter"
USING (permission_check(resource_path, 'notification_class_filter'))
WITH CHECK (permission_check(resource_path, 'notification_class_filter'));

CREATE POLICY rls_notification_class_filter_restrictive ON "notification_class_filter" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'notification_class_filter'))
with check (permission_check(resource_path, 'notification_class_filter'));

ALTER TABLE "notification_class_filter" ENABLE ROW LEVEL security;
ALTER TABLE "notification_class_filter" FORCE ROW LEVEL security;
