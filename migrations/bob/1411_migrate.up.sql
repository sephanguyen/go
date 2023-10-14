-- Withus -2147483629 --
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GGVXWS2AMVMCXWJWV1XEFHDR', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629'),
  ('01GGVXWS2HF8G3GFTVYEFSTBPR', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629'),
  ('01GGVXWRZTN0N5ENQEV399KQH1', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;
INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S1')
ON CONFLICT ON CONSTRAINT granted_permission__pk
DO UPDATE SET user_group_name = excluded.user_group_name;

-- Withus -2147483630 --
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GGVRH1PNSTJTNZJ3MZBFDXKT', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630'),
  ('01GGVRH1P9HZQ7T4EX424B1BCA', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630'),
  ('01GGVRH1J0B91MSN1AHPT6YMBB', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;
INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S2')
ON CONFLICT ON CONSTRAINT granted_permission__pk
DO UPDATE SET user_group_name = excluded.user_group_name;

-- Eshinkan -2147483631 --
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GGVJB2DCDK9V0TC3JWRCHES5', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631'),
  ('01GGVJB2CJZ9JEN4P309WJ1PSA', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631'),
  ('01GGVJB2BXQ0WBW6FXPKTKP9K9', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;
INSERT INTO granted_permission
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S3')
ON CONFLICT ON CONSTRAINT granted_permission__pk
DO UPDATE SET user_group_name = excluded.user_group_name;
