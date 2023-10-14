DROP POLICY IF EXISTS rls_info_notifications_update_location on "info_notifications";

CREATE POLICY rls_info_notifications_update_location ON "info_notifications" AS PERMISSIVE FOR
update
    TO PUBLIC using (
        true <= (
            select
                true
            from
                granted_permissions p
                join info_notifications_access_paths inap on inap.location_id = p.location_id
            where
                p.user_id = current_setting('app.user_id')
                and p.permission_id = (
                    select
                        p2.permission_id
                    from
                        permission p2
                    where
                        p2.permission_name = 'communication.notification.read'
                        and p2.resource_path = current_setting('permission.resource_path')
                )
                and inap."notification_id" = info_notifications.notification_id
                and inap."deleted_at" is null
            limit
                1
        )
    ) with check (
        true <= (
            select
                true
            from
                granted_permissions p
                join info_notifications_access_paths inap on inap.location_id = p.location_id
            where
                p.user_id = current_setting('app.user_id')
                and p.permission_id = (
                    select
                        p2.permission_id
                    from
                        permission p2
                    where
                        p2.permission_name = 'communication.notification.write'
                        and p2.resource_path = current_setting('permission.resource_path')
                )
                and inap."notification_id" = info_notifications.notification_id
                and inap."deleted_at" is null
            limit
                1
        )
    );
