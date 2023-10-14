DROP POLICY IF EXISTS rls_locations on "locations";
CREATE POLICY rls_locations_location ON "locations" AS PERMISSIVE FOR all TO PUBLIC
using (location_id in (
  select
    p.location_id
  from
          granted_permissions p
  where
    p.user_id = current_setting('app.user_id')
    and p.permission_id = (
      select
        p2.permission_id
      from
        "permission" p2
      where
        p2.permission_name = 'master.location.read'
        and p2.resource_path = current_setting('permission.resource_path'))
  )
)
with check (exists (
  select
    true
  from
          granted_permissions p
  where
    p.user_id = current_setting('app.user_id')
    and p.permission_id = (
      select
        p2.permission_id
      from
        "permission" p2
      where
        p2.permission_name = 'master.location.write'
        and p2.resource_path = current_setting('permission.resource_path'))
  )
);