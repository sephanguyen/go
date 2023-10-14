DROP POLICY IF EXISTS rls_info_notifications_access_paths on "info_notifications_access_paths";

DROP POLICY IF EXISTS rls_info_notifications_access_paths_location_permission_v4 on "info_notifications_access_paths";

CREATE POLICY rls_info_notifications_access_paths_location_permission_v4 ON "info_notifications_access_paths" AS PERMISSIVE FOR all TO PUBLIC using (
    true <= (
        select
            true
        from
            granted_permissions p
        where
            p.user_id = current_setting('app.user_id')
            and p.location_id = info_notifications_access_paths.location_id
            and (
                p.permission_id = (
                    select
                        p2.permission_id
                    from
                        permission p2
                    where
                        p2.permission_name = 'communication.notification.read'
                        and p2.resource_path = current_setting('permission.resource_path')
                )
                or (
                    p.permission_id = (
                        select
                            p2.permission_id
                        from
                            permission p2
                        where
                            p2.permission_name = 'communication.notification.owner'
                            and p2.resource_path = current_setting('permission.resource_path')
                    )
                    and info_notifications_access_paths.created_user_id = current_setting('app.user_id')
                )
            )
        limit
            1
    )
) with check (
    true <= (
        select
            true
        from
            granted_permissions p
        where
            p.user_id = current_setting('app.user_id')
            and p.location_id = info_notifications_access_paths.location_id
            and (
                p.permission_id = (
                    select
                        p2.permission_id
                    from
                        permission p2
                    where
                        p2.permission_name = 'communication.notification.write'
                        and p2.resource_path = current_setting('permission.resource_path')
                )
                or (
                    p.permission_id = (
                        select
                            p2.permission_id
                        from
                            permission p2
                        where
                            p2.permission_name = 'communication.notification.owner'
                            and p2.resource_path = current_setting('permission.resource_path')
                    )
                    and info_notifications_access_paths.created_user_id = current_setting('app.user_id')
                )
            )
        limit
            1
    )
);