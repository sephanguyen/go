DROP POLICY IF EXISTS rls_ac_test_template_11_4_select_location on ac_test_template_11_4;
DROP POLICY IF EXISTS rls_ac_test_template_11_4_insert_location on ac_test_template_11_4;
DROP POLICY IF EXISTS rls_ac_test_template_11_4_update_location on ac_test_template_11_4;
DROP POLICY IF EXISTS rls_ac_test_template_11_4_delete_location on ac_test_template_11_4;

CREATE POLICY rls_ac_test_template_11_4_select_location ON ac_test_template_11_4 AS PERMISSIVE FOR select TO PUBLIC
using (true <= (
  select			
    true
  from
          granted_permissions p
  join ac_test_template_11_4_access_paths usp on
          usp.location_id = p.location_id
  where
    p.user_id = current_setting('app.user_id')
    and p.permission_name = 'accesscontrol.ac_test_template_11_4.read'
    and usp."ac_test_template_11_4_id" = ac_test_template_11_4.ac_test_template_11_4_id
  limit 1
  )
)
;CREATE POLICY rls_ac_test_template_11_4_insert_location ON ac_test_template_11_4 AS PERMISSIVE FOR insert TO PUBLIC

with check ((
  1 = 1
)
);CREATE POLICY rls_ac_test_template_11_4_update_location ON ac_test_template_11_4 AS PERMISSIVE FOR update TO PUBLIC
using (true <= (
  select			
    true
  from
          granted_permissions p
  join ac_test_template_11_4_access_paths usp on
          usp.location_id = p.location_id
  where
    p.user_id = current_setting('app.user_id')
    and p.permission_name = 'accesscontrol.ac_test_template_11_4.write'
    and usp."ac_test_template_11_4_id" = ac_test_template_11_4.ac_test_template_11_4_id
  limit 1
  )
)
with check (true <= (
  select			
    true
  from
          granted_permissions p
  join ac_test_template_11_4_access_paths usp on
          usp.location_id = p.location_id
  where
    p.user_id = current_setting('app.user_id')
    and p.permission_name = 'accesscontrol.ac_test_template_11_4.write'
    and usp."ac_test_template_11_4_id" = ac_test_template_11_4.ac_test_template_11_4_id
  limit 1
  )
);CREATE POLICY rls_ac_test_template_11_4_delete_location ON ac_test_template_11_4 AS PERMISSIVE FOR delete TO PUBLIC
using (true <= (
  select			
    true
  from
          granted_permissions p
  join ac_test_template_11_4_access_paths usp on
          usp.location_id = p.location_id
  where
    p.user_id = current_setting('app.user_id')
    and p.permission_name = 'accesscontrol.ac_test_template_11_4.write'
    and usp."ac_test_template_11_4_id" = ac_test_template_11_4.ac_test_template_11_4_id
  limit 1
  )
)
;