CREATE TABLE IF NOT EXISTS public.notification_date (
                                                        notification_date_id text NOT NULL,
                                                        order_type text NOT NULL,
                                                        notification_date INT NOT NULL,
                                                        is_archived boolean DEFAULT FALSE NOT NULL,
                                                        created_at timestamp with time zone NOT NULL,
                                                        updated_at timestamp with time zone NOT NULL,
                                                        resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT notification_date__notification_date_id__pk PRIMARY KEY (notification_date_id)
    );

CREATE POLICY rls_notification_date ON "notification_date"
    USING (permission_check(resource_path, 'notification_date'))
    WITH CHECK (permission_check(resource_path, 'notification_date'));

CREATE POLICY rls_notification_date_restrictive ON "notification_date"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'notification_date'))
    WITH CHECK (permission_check(resource_path, 'notification_date'));

ALTER TABLE "notification_date" ENABLE ROW LEVEL security;
ALTER TABLE "notification_date" FORCE ROW LEVEL security;
