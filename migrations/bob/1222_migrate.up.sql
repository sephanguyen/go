CREATE TABLE IF NOT EXISTS public.tags (
    tag_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk__tags PRIMARY KEY (tag_id)
);

CREATE TABLE IF NOT EXISTS public.info_notifications_tags (
    notification_tag_id TEXT NOT NULL,
    notification_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk__notifications_tags PRIMARY KEY (notification_tag_id),
	CONSTRAINT fk__notifications_tags__notification_id FOREIGN KEY(notification_id) REFERENCES public.info_notifications(notification_id),
	CONSTRAINT fk__notifications_tags__tag_id FOREIGN KEY(tag_id) REFERENCES public.tags(tag_id)
);


CREATE POLICY rls_tags ON "tags"
USING (permission_check(resource_path, 'tags'))
WITH CHECK (permission_check(resource_path, 'tags'));

ALTER TABLE "tags" ENABLE ROW LEVEL security;
ALTER TABLE "tags" FORCE ROW LEVEL security;

CREATE POLICY rls_info_notifications_tags ON "info_notifications_tags"
USING (permission_check(resource_path, 'info_notifications_tags'))
WITH CHECK (permission_check(resource_path, 'info_notifications_tags'));

ALTER TABLE "info_notifications_tags" ENABLE ROW LEVEL security;
ALTER TABLE "info_notifications_tags" FORCE ROW LEVEL security;
