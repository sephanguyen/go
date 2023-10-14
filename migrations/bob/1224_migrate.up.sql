-- migration for resource_path -2147483637
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01G6A5YH5MS80ZEPS9CSNW8KPP', 'master.location.read', now(), now(), '-2147483637')
	ON CONFLICT DO NOTHING;

INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01G6A5WBRD3B447Z2SVDQVAGJ2', 'Teacher', false, now(), now(), '-2147483637'),
  ('01G6A5WBRD3B447Z2SVDQVAGJ3', 'School Admin', false, now(), now(), '-2147483637'),
  ('01G6A5Z2SWEN7GJ2DTT4BJ38X1', 'Student', true, now(), now(), '-2147483637'),
  ('01G6A5Z2SWEN7GJ2DTT4BJ38X2', 'Parent',  true, now(), now(), '-2147483637')
	ON CONFLICT DO NOTHING;

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01G6A5YH5MS80ZEPS9CSNW8KPP', '01G6A5WBRD3B447Z2SVDQVAGJ2', now(), now(), '-2147483637'),
  ('01G6A5YH5MS80ZEPS9CSNW8KPP', '01G6A5WBRD3B447Z2SVDQVAGJ3', now(), now(), '-2147483637'),
  ('01G6A5YH5MS80ZEPS9CSNW8KPP', '01G6A5Z2SWEN7GJ2DTT4BJ38X1', now(), now(), '-2147483637'),
  ('01G6A5YH5MS80ZEPS9CSNW8KPP', '01G6A5Z2SWEN7GJ2DTT4BJ38X2', now(), now(), '-2147483637')
	ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01G6A5ZMAWPNQJZQQYA0FRNFB1', 'Student', true, now(), now(), '-2147483637'),
  ('01G6A5ZMAWPNQJZQQYA0FRNFB2', 'Parent',  true, now(), now(), '-2147483637')

  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01G6A5Y1DXMFEKWJNQ9G9JPG21', '01G6A5ZMAWPNQJZQQYA0FRNFB1', '01G6A5Z2SWEN7GJ2DTT4BJ38X1', now(), now(), '-2147483637'),
  ('01G6A5Y1DXMFEKWJNQ9G9JPG22', '01G6A5ZMAWPNQJZQQYA0FRNFB2', '01G6A5Z2SWEN7GJ2DTT4BJ38X2', now(), now(), '-2147483637')

  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01G6A5Y1DXMFEKWJNQ9G9JPG21', '01FR4M51XJY9E77GSN4QZ1Q8N3', now(), now(), '-2147483637'),
  ('01G6A5Y1DXMFEKWJNQ9G9JPG22', '01FR4M51XJY9E77GSN4QZ1Q8N3', now(), now(), '-2147483637')

  ON CONFLICT DO NOTHING;

-- migration for resource_path -2147483638
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01G6A69MMBG2A02KRDQE1F2QPT', 'master.location.read', now(), now(), '-2147483638')
	ON CONFLICT DO NOTHING;

INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01G6A6B3YP6P161SH6CJN3QPF0', 'Teacher', false, now(), now(), '-2147483638'),
  ('01G6A6B3YP6P161SH6CJN3QPF1', 'School Admin', false, now(), now(), '-2147483638'),
  ('01G6A69WTMFK8NW10FJCNAS530', 'Student', true, now(), now(), '-2147483638'),
  ('01G6A69WTMFK8NW10FJCNAS531', 'Parent',  true, now(), now(), '-2147483638')
	ON CONFLICT DO NOTHING;

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01G6A69MMBG2A02KRDQE1F2QPT', '01G6A6B3YP6P161SH6CJN3QPF0', now(), now(), '-2147483638'),
  ('01G6A69MMBG2A02KRDQE1F2QPT', '01G6A6B3YP6P161SH6CJN3QPF1', now(), now(), '-2147483638'),
  ('01G6A69MMBG2A02KRDQE1F2QPT', '01G6A69WTMFK8NW10FJCNAS530', now(), now(), '-2147483638'),
  ('01G6A69MMBG2A02KRDQE1F2QPT', '01G6A69WTMFK8NW10FJCNAS531', now(), now(), '-2147483638')
	ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01G6A6A7FX5NP58RFBXR7V4RD0', 'Student', true, now(), now(), '-2147483638'),
  ('01G6A6A7FX5NP58RFBXR7V4RD1', 'Parent',  true, now(), now(), '-2147483638')

  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01G6A6APJJ1NWSQMVAWK1311A0', '01G6A6A7FX5NP58RFBXR7V4RD0', '01G6A69WTMFK8NW10FJCNAS530', now(), now(), '-2147483638'),
  ('01G6A6APJJ1NWSQMVAWK1311A1', '01G6A6A7FX5NP58RFBXR7V4RD1', '01G6A69WTMFK8NW10FJCNAS531', now(), now(), '-2147483638')

  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01G6A6APJJ1NWSQMVAWK1311A0', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638'),
  ('01G6A6APJJ1NWSQMVAWK1311A1', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638')

  ON CONFLICT DO NOTHING;
