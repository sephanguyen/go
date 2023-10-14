
-- (-2147483648) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXR1EATFX5SVZPYG6E6', 'ChatSystemAdmin', true, now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCARWA5MS7QFW12A365R4B0', '01H7FDSCXR1EATFX5SVZPYG6E6', now(), now(), '-2147483648'),
    ('01GDCARWA5MS7QFW12A365R4B1', '01H7FDSCXR1EATFX5SVZPYG6E6', now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXRZM21RYYCKZXJXRMF', 'ChatSystemAdmin', true, now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSSXKDQBG62W6SQDP0', '01H7FDSCXRZM21RYYCKZXJXRMF', '01H7FDSCXR1EATFX5SVZPYG6E6', now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSSXKDQBG62W6SQDP0', '01FR4M51XJY9E77GSN4QZ1Q9N1', now(), now(), '-2147483648')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXSR7WHVHCQ5SCC62KE', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXSR7WHVHCQ5SCC62KE', '01H7FDSCXRZM21RYYCKZXJXRMF', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N1') 
ON CONFLICT DO NOTHING;
        
-- (-2147483647) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXSGKRQC1FNHRXZHRST', 'ChatSystemAdmin', true, now(), now(), '-2147483647')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAQGQCD0Y2XWQ7BYQH2F00', '01H7FDSCXSGKRQC1FNHRXZHRST', now(), now(), '-2147483647'),
    ('01GDCAQGQCD0Y2XWQ7BYQH2F01', '01H7FDSCXSGKRQC1FNHRXZHRST', now(), now(), '-2147483647')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXS5090ZKS8KJQSHCXM', 'ChatSystemAdmin', true, now(), now(), '-2147483647')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSG1EQWC0YJSB89YR5', '01H7FDSCXS5090ZKS8KJQSHCXM', '01H7FDSCXSGKRQC1FNHRXZHRST', now(), now(), '-2147483647')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSG1EQWC0YJSB89YR5', '01FR4M51XJY9E77GSN4QZ1Q9N2', now(), now(), '-2147483647')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXSP5AEEXRCHTW84BHC', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483647', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXSP5AEEXRCHTW84BHC', '01H7FDSCXS5090ZKS8KJQSHCXM', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N2') 
ON CONFLICT DO NOTHING;
        
-- (-2147483646) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXS5PPNFF5B7RJ6QYEZ', 'ChatSystemAdmin', true, now(), now(), '-2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAP2C1CTSYGQM188YMT0X0', '01H7FDSCXS5PPNFF5B7RJ6QYEZ', now(), now(), '-2147483646'),
    ('01GDCAP2C1CTSYGQM188YMT0X1', '01H7FDSCXS5PPNFF5B7RJ6QYEZ', now(), now(), '-2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSDKVED8QYHQBKDY1H', 'ChatSystemAdmin', true, now(), now(), '-2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSERZ301PMZJHD01MN', '01H7FDSCXSDKVED8QYHQBKDY1H', '01H7FDSCXS5PPNFF5B7RJ6QYEZ', now(), now(), '-2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSERZ301PMZJHD01MN', '01FR4M51XJY9E77GSN4QZ1Q9N3', now(), now(), '-2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXSZE0YMG5GS22RC9VZ', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483646', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXSZE0YMG5GS22RC9VZ', '01H7FDSCXSDKVED8QYHQBKDY1H', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N3') 
ON CONFLICT DO NOTHING;
        
-- (2147483646) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXSY500ZRKPRA8B6E1V', 'ChatSystemAdmin', true, now(), now(), '2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4JC53MX00QMQ3HM02GHZBR4', '01H7FDSCXSY500ZRKPRA8B6E1V', now(), now(), '2147483646'),
    ('01H4JC53MY01KKZSVS022CYSQC', '01H7FDSCXSY500ZRKPRA8B6E1V', now(), now(), '2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSG36JHGHJZJCJ7F42', 'ChatSystemAdmin', true, now(), now(), '2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSCJS01VR97CKSG2QM', '01H7FDSCXSG36JHGHJZJCJ7F42', '01H7FDSCXSY500ZRKPRA8B6E1V', now(), now(), '2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSCJS01VR97CKSG2QM', '01H4JC53MQ00WXEHV200VS2JKZ', now(), now(), '2147483646')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXSMEF22D0E1VVR6F47', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483646', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXSMEF22D0E1VVR6F47', '01H7FDSCXSG36JHGHJZJCJ7F42', now(), now(), '01H4JC53MQ00WXEHV200VS2JKZ') 
ON CONFLICT DO NOTHING;
        
-- (-2147483645) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXS3EG35WNHRMZ42NYK', 'ChatSystemAdmin', true, now(), now(), '-2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAMQHAEGJGPY019KR7XRC0', '01H7FDSCXS3EG35WNHRMZ42NYK', now(), now(), '-2147483645'),
    ('01GDCAMQHAEGJGPY019KR7XRC1', '01H7FDSCXS3EG35WNHRMZ42NYK', now(), now(), '-2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSGRXVJ7PTVWVB2JX7', 'ChatSystemAdmin', true, now(), now(), '-2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXS4HWKSRZ5W4SVWFHD', '01H7FDSCXSGRXVJ7PTVWVB2JX7', '01H7FDSCXS3EG35WNHRMZ42NYK', now(), now(), '-2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXS4HWKSRZ5W4SVWFHD', '01FR4M51XJY9E77GSN4QZ1Q9N4', now(), now(), '-2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXSS6CRAAKG6SQTBYW7', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483645', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXSS6CRAAKG6SQTBYW7', '01H7FDSCXSGRXVJ7PTVWVB2JX7', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N4') 
ON CONFLICT DO NOTHING;
        
-- (2147483645) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXSGHWT6ZTW8KRGRMJB', 'ChatSystemAdmin', true, now(), now(), '2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4QBARW1018YSKCM01PK6S5R', '01H7FDSCXSGHWT6ZTW8KRGRMJB', now(), now(), '2147483645'),
    ('01H4QBARW000TBTK4G02SFWGQ7', '01H7FDSCXSGHWT6ZTW8KRGRMJB', now(), now(), '2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXST7TTP82P99M5G95V', 'ChatSystemAdmin', true, now(), now(), '2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSSVFTE7CX59TNBP0R', '01H7FDSCXST7TTP82P99M5G95V', '01H7FDSCXSGHWT6ZTW8KRGRMJB', now(), now(), '2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXSSVFTE7CX59TNBP0R', '01H4QBARVT01E3GNWY00KG5212', now(), now(), '2147483645')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXSX305NWEPQWGDFRRC', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483645', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXSX305NWEPQWGDFRRC', '01H7FDSCXST7TTP82P99M5G95V', now(), now(), '01H4QBARVT01E3GNWY00KG5212') 
ON CONFLICT DO NOTHING;
        
-- (-2147483644) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXSXDA6SXKG2HXJBWYK', 'ChatSystemAdmin', true, now(), now(), '-2147483644')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAK4EK5CESWF55RA7BT201', '01H7FDSCXSXDA6SXKG2HXJBWYK', now(), now(), '-2147483644'),
    ('01GDCAK4EK5CESWF55RA7BT200', '01H7FDSCXSXDA6SXKG2HXJBWYK', now(), now(), '-2147483644')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTTVHVM9AT07XSMP67', 'ChatSystemAdmin', true, now(), now(), '-2147483644')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTRN72H6VWY4KKCJ4G', '01H7FDSCXTTVHVM9AT07XSMP67', '01H7FDSCXSXDA6SXKG2HXJBWYK', now(), now(), '-2147483644')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTRN72H6VWY4KKCJ4G', '01FR4M51XJY9E77GSN4QZ1Q9N5', now(), now(), '-2147483644')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXTTYFS0B1X3E0GW6JP', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483644', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXTTYFS0B1X3E0GW6JP', '01H7FDSCXTTVHVM9AT07XSMP67', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N5') 
ON CONFLICT DO NOTHING;
        
-- (-2147483643) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXTJN5653V9DEF2QANT', 'ChatSystemAdmin', true, now(), now(), '-2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAHJBCA1RWVFTC861929A2', '01H7FDSCXTJN5653V9DEF2QANT', now(), now(), '-2147483643'),
    ('01GDCAHJBCA1RWVFTC861929A1', '01H7FDSCXTJN5653V9DEF2QANT', now(), now(), '-2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTW6KG04XZVX7R3X1H', 'ChatSystemAdmin', true, now(), now(), '-2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXT3XT0040FZD44WG2M', '01H7FDSCXTW6KG04XZVX7R3X1H', '01H7FDSCXTJN5653V9DEF2QANT', now(), now(), '-2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXT3XT0040FZD44WG2M', '01FR4M51XJY9E77GSN4QZ1Q9N6', now(), now(), '-2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXTQKV7XJ5YTH4VPHZF', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483643', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXTQKV7XJ5YTH4VPHZF', '01H7FDSCXTW6KG04XZVX7R3X1H', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N6') 
ON CONFLICT DO NOTHING;
        
-- (2147483643) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXT9ZAKAT19VJE33DFH', 'ChatSystemAdmin', true, now(), now(), '2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4QG7EKQ01NZN71H02JPX3P5', '01H7FDSCXT9ZAKAT19VJE33DFH', now(), now(), '2147483643'),
    ('01H4QG7EKR009DG7XF02M7Z5RR', '01H7FDSCXT9ZAKAT19VJE33DFH', now(), now(), '2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTDQ75VCTMFNB1BVJ0', 'ChatSystemAdmin', true, now(), now(), '2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTPVB8H9FCZDQ3VYQ8', '01H7FDSCXTDQ75VCTMFNB1BVJ0', '01H7FDSCXT9ZAKAT19VJE33DFH', now(), now(), '2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTPVB8H9FCZDQ3VYQ8', '01H4QG7EKN017PBSEA021WZPZ4', now(), now(), '2147483643')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXTGGPE27H5B24TR18Q', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483643', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXTGGPE27H5B24TR18Q', '01H7FDSCXTDQ75VCTMFNB1BVJ0', now(), now(), '01H4QG7EKN017PBSEA021WZPZ4') 
ON CONFLICT DO NOTHING;
        
-- (-2147483642) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXT6HQ7FS4AWGQ1N8ZP', 'ChatSystemAdmin', true, now(), now(), '-2147483642')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAFXAN6N6YYTYWHHJZ5EP1', '01H7FDSCXT6HQ7FS4AWGQ1N8ZP', now(), now(), '-2147483642'),
    ('01GDCAFXAN6N6YYTYWHHJZ5EP2', '01H7FDSCXT6HQ7FS4AWGQ1N8ZP', now(), now(), '-2147483642')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXTRXE7BA9ZTHF6Q9BT', 'ChatSystemAdmin', true, now(), now(), '-2147483642')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVM2ZMP5QXMX1W4TNH', '01H7FDSCXTRXE7BA9ZTHF6Q9BT', '01H7FDSCXT6HQ7FS4AWGQ1N8ZP', now(), now(), '-2147483642')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVM2ZMP5QXMX1W4TNH', '01FR4M51XJY9E77GSN4QZ1Q9N7', now(), now(), '-2147483642')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXVSZTQZR998AYJHVAH', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483642', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXVSZTQZR998AYJHVAH', '01H7FDSCXTRXE7BA9ZTHF6Q9BT', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N7') 
ON CONFLICT DO NOTHING;
        
-- (-2147483641) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXVNEAGJDKEPHE355GP', 'ChatSystemAdmin', true, now(), now(), '-2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAEE66NTD5NA3C2BQR59D0', '01H7FDSCXVNEAGJDKEPHE355GP', now(), now(), '-2147483641'),
    ('01GDCAEE66NTD5NA3C2BQR59D1', '01H7FDSCXVNEAGJDKEPHE355GP', now(), now(), '-2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVXN1AQG62R387BM1A', 'ChatSystemAdmin', true, now(), now(), '-2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVF1SQR7AKDPD1VMPM', '01H7FDSCXVXN1AQG62R387BM1A', '01H7FDSCXVNEAGJDKEPHE355GP', now(), now(), '-2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVF1SQR7AKDPD1VMPM', '01FR4M51XJY9E77GSN4QZ1Q9N8', now(), now(), '-2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXVWW8Y9YYY0D1FJ43W', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483641', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXVWW8Y9YYY0D1FJ43W', '01H7FDSCXVXN1AQG62R387BM1A', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N8') 
ON CONFLICT DO NOTHING;
        
-- (2147483641) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXVRKZF78RA5Z2KE4RE', 'ChatSystemAdmin', true, now(), now(), '2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4QBRYSZ00F9VPG10282Q0EX', '01H7FDSCXVRKZF78RA5Z2KE4RE', now(), now(), '2147483641'),
    ('01H4QBRYSY03Z1QJ7801ZZ1VTW', '01H7FDSCXVRKZF78RA5Z2KE4RE', now(), now(), '2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVWEAE4YJHRQDNXDDC', 'ChatSystemAdmin', true, now(), now(), '2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVK7R99BS7A4JP65TQ', '01H7FDSCXVWEAE4YJHRQDNXDDC', '01H7FDSCXVRKZF78RA5Z2KE4RE', now(), now(), '2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVK7R99BS7A4JP65TQ', '01H4QBRYSW00VRWAF80182W5KB', now(), now(), '2147483641')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXVDX5NRNY7Q3M78SCD', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483641', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXVDX5NRNY7Q3M78SCD', '01H7FDSCXVWEAE4YJHRQDNXDDC', now(), now(), '01H4QBRYSW00VRWAF80182W5KB') 
ON CONFLICT DO NOTHING;
        
-- (-2147483640) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXVZDYVTTNH9PWHJ2DD', 'ChatSystemAdmin', true, now(), now(), '-2147483640')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCAD4CDE87P7V59XGAXDHW2', '01H7FDSCXVZDYVTTNH9PWHJ2DD', now(), now(), '-2147483640'),
    ('01GDCAD4CDE87P7V59XGAXDHW1', '01H7FDSCXVZDYVTTNH9PWHJ2DD', now(), now(), '-2147483640')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXVJRRE5P9VMWWGJ3D9', 'ChatSystemAdmin', true, now(), now(), '-2147483640')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXV160XSA2H205SBJVG', '01H7FDSCXVJRRE5P9VMWWGJ3D9', '01H7FDSCXVZDYVTTNH9PWHJ2DD', now(), now(), '-2147483640')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXV160XSA2H205SBJVG', '01FR4M51XJY9E77GSN4QZ1Q9N9', now(), now(), '-2147483640')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXVZG6V0WEYW98Q0J77', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483640', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXVZG6V0WEYW98Q0J77', '01H7FDSCXVJRRE5P9VMWWGJ3D9', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q9N9') 
ON CONFLICT DO NOTHING;
        
-- (-2147483639) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXVHDG8AAGP4FC28R20', 'ChatSystemAdmin', true, now(), now(), '-2147483639')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCA61W0M2X13215JM6JWKQ1', '01H7FDSCXVHDG8AAGP4FC28R20', now(), now(), '-2147483639'),
    ('01GDCA61W0M2X13215JM6JWKQ2', '01H7FDSCXVHDG8AAGP4FC28R20', now(), now(), '-2147483639')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWFMG8THVFAVWJ0PV9', 'ChatSystemAdmin', true, now(), now(), '-2147483639')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWQPKGNBF5G0Z96GG7', '01H7FDSCXWFMG8THVFAVWJ0PV9', '01H7FDSCXVHDG8AAGP4FC28R20', now(), now(), '-2147483639')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWQPKGNBF5G0Z96GG7', '01FR4M51XJY9E77GSN4QZ1Q8N1', now(), now(), '-2147483639')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXWDT3T9KQ9EACN8QXZ', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483639', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXWDT3T9KQ9EACN8QXZ', '01H7FDSCXWFMG8THVFAVWJ0PV9', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q8N1') 
ON CONFLICT DO NOTHING;
        
-- (-2147483638) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXWFBY0F0DNR9DD217Y', 'ChatSystemAdmin', true, now(), now(), '-2147483638')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCA3WFW1P5MFGM9MVGFW694', '01H7FDSCXWFBY0F0DNR9DD217Y', now(), now(), '-2147483638'),
    ('01GDCA3WFW1P5MFGM9MVGFW693', '01H7FDSCXWFBY0F0DNR9DD217Y', now(), now(), '-2147483638')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWYHY1SW96BDP9DQD1', 'ChatSystemAdmin', true, now(), now(), '-2147483638')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXW660J2TQ02Z5PSW16', '01H7FDSCXWYHY1SW96BDP9DQD1', '01H7FDSCXWFBY0F0DNR9DD217Y', now(), now(), '-2147483638')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXW660J2TQ02Z5PSW16', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXWV6WHVA2ZFJPK5GB2', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483638', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXWV6WHVA2ZFJPK5GB2', '01H7FDSCXWYHY1SW96BDP9DQD1', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q8N2') 
ON CONFLICT DO NOTHING;
        
-- (-2147483637) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXWWPX364J33VRE1CEB', 'ChatSystemAdmin', true, now(), now(), '-2147483637')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDCA26BVQS9SC9M8HC3Z8TV2', '01H7FDSCXWWPX364J33VRE1CEB', now(), now(), '-2147483637'),
    ('01GDCA26BVQS9SC9M8HC3Z8TV1', '01H7FDSCXWWPX364J33VRE1CEB', now(), now(), '-2147483637')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWVM67Z3784ZXYR10S', 'ChatSystemAdmin', true, now(), now(), '-2147483637')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWTG3M77GFHVNNVTKN', '01H7FDSCXWVM67Z3784ZXYR10S', '01H7FDSCXWWPX364J33VRE1CEB', now(), now(), '-2147483637')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWTG3M77GFHVNNVTKN', '01FR4M51XJY9E77GSN4QZ1Q8N3', now(), now(), '-2147483637')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXWW1Y3A76ZAEND3GK5', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483637', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXWW1Y3A76ZAEND3GK5', '01H7FDSCXWVM67Z3784ZXYR10S', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q8N3') 
ON CONFLICT DO NOTHING;
        
-- (-2147483635) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXWSZ2YMHY3WHE6JFG3', 'ChatSystemAdmin', true, now(), now(), '-2147483635')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDC9Y7V9RPXVZ5N5PH983AJ3', '01H7FDSCXWSZ2YMHY3WHE6JFG3', now(), now(), '-2147483635'),
    ('01GDC9Y7V9RPXVZ5N5PH983AJ2', '01H7FDSCXWSZ2YMHY3WHE6JFG3', now(), now(), '-2147483635')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWHAA0VZPQF60X5WH9', 'ChatSystemAdmin', true, now(), now(), '-2147483635')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWGF0ESTVFBGKBHMNQ', '01H7FDSCXWHAA0VZPQF60X5WH9', '01H7FDSCXWSZ2YMHY3WHE6JFG3', now(), now(), '-2147483635')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWGF0ESTVFBGKBHMNQ', '01FR4M51XJY9E77GSN4QZ1Q8N4', now(), now(), '-2147483635')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXW6SCMP94S5V8J1HFD', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483635', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXW6SCMP94S5V8J1HFD', '01H7FDSCXWHAA0VZPQF60X5WH9', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q8N4') 
ON CONFLICT DO NOTHING;
        
-- (-2147483634) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXWB5NKFN26VMM1P5JM', 'ChatSystemAdmin', true, now(), now(), '-2147483634')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GDC9V6C776PQH7FP4VMC6JD2', '01H7FDSCXWB5NKFN26VMM1P5JM', now(), now(), '-2147483634'),
    ('01GDC9V6C776PQH7FP4VMC6JD1', '01H7FDSCXWB5NKFN26VMM1P5JM', now(), now(), '-2147483634')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXWXTAC2CSN0KW33XK0', 'ChatSystemAdmin', true, now(), now(), '-2147483634')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXKKDTP0BAK7X32HYP', '01H7FDSCXWXTAC2CSN0KW33XK0', '01H7FDSCXWB5NKFN26VMM1P5JM', now(), now(), '-2147483634')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXKKDTP0BAK7X32HYP', '01FR4M51XJY9E77GSN4QZ1Q8N5', now(), now(), '-2147483634')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXXD2HQZQ3KMC1BP83S', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483634', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXXD2HQZQ3KMC1BP83S', '01H7FDSCXWXTAC2CSN0KW33XK0', now(), now(), '01FR4M51XJY9E77GSN4QZ1Q8N5') 
ON CONFLICT DO NOTHING;
        
-- (-2147483631) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXXDXM0BHRKD3ZC3EAH', 'ChatSystemAdmin', true, now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GGVJB2C7935GH4DE6MSEQ4T7', '01H7FDSCXXDXM0BHRKD3ZC3EAH', now(), now(), '-2147483631'),
    ('01GGVJB2C8PXABWY151S8HS8W1', '01H7FDSCXXDXM0BHRKD3ZC3EAH', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXZVKG0CPPDRY28NFE', 'ChatSystemAdmin', true, now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXX0RF4CT57DW1XA7R0', '01H7FDSCXXZVKG0CPPDRY28NFE', '01H7FDSCXXDXM0BHRKD3ZC3EAH', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXX0RF4CT57DW1XA7R0', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXXFDNMDGTT17PBTWJE', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483631', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXXFDNMDGTT17PBTWJE', '01H7FDSCXXZVKG0CPPDRY28NFE', now(), now(), '01GDWSMJS6APH4SX2NP5NFWHG5') 
ON CONFLICT DO NOTHING;
        
-- (2147483631) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXXCYAQXMJWTN05FHYG', 'ChatSystemAdmin', true, now(), now(), '2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4QGKSC800NTP0HE013ZDYS8', '01H7FDSCXXCYAQXMJWTN05FHYG', now(), now(), '2147483631'),
    ('01H4QGKSC702WBZKS403H01CJ1', '01H7FDSCXXCYAQXMJWTN05FHYG', now(), now(), '2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXNH7ZQTAN77S7GARH', 'ChatSystemAdmin', true, now(), now(), '2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXRWNAVEQ6XHAZ2BNV', '01H7FDSCXXNH7ZQTAN77S7GARH', '01H7FDSCXXCYAQXMJWTN05FHYG', now(), now(), '2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXRWNAVEQ6XHAZ2BNV', '01H4QGKSC500SKGEBX03YVYZM5', now(), now(), '2147483631')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXX0KXJZH6JD1SW8XP7', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483631', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXX0KXJZH6JD1SW8XP7', '01H7FDSCXXNH7ZQTAN77S7GARH', now(), now(), '01H4QGKSC500SKGEBX03YVYZM5') 
ON CONFLICT DO NOTHING;
        
-- (-2147483630) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXXQJBWXTGH12CAA8ED', 'ChatSystemAdmin', true, now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GGVRH1N3ZMGD96C31PSVSSVX', '01H7FDSCXXQJBWXTGH12CAA8ED', now(), now(), '-2147483630'),
    ('01GGVRH1N6VB30XBB26PGY562C', '01H7FDSCXXQJBWXTGH12CAA8ED', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXX7WBC2Z9V1JYKGVER', 'ChatSystemAdmin', true, now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXNFKJDRB538ZMPF2E', '01H7FDSCXX7WBC2Z9V1JYKGVER', '01H7FDSCXXQJBWXTGH12CAA8ED', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXXNFKJDRB538ZMPF2E', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXXN7SFGC6PEHBJZA5G', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483630', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXXN7SFGC6PEHBJZA5G', '01H7FDSCXX7WBC2Z9V1JYKGVER', now(), now(), '01GFMMFRXC6SKTTT44HWR3BRY8') 
ON CONFLICT DO NOTHING;
        
-- (2147483630) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXXT60SBHESK4TZZW7Y', 'ChatSystemAdmin', true, now(), now(), '2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4Z1V2X903RH7RH500H3CCTH', '01H7FDSCXXT60SBHESK4TZZW7Y', now(), now(), '2147483630'),
    ('01H4Z1V2XA02VZANTP01XCBWFX', '01H7FDSCXXT60SBHESK4TZZW7Y', now(), now(), '2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXX2PJ4GNPE2V5GR0EC', 'ChatSystemAdmin', true, now(), now(), '2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYFS3RNKSXZGF6EZ12', '01H7FDSCXX2PJ4GNPE2V5GR0EC', '01H7FDSCXXT60SBHESK4TZZW7Y', now(), now(), '2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYFS3RNKSXZGF6EZ12', '01H4Z1V2X401KBR5B102N9ZJ4T', now(), now(), '2147483630')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXYB74TSPXWWP2HPS7M', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483630', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXYB74TSPXWWP2HPS7M', '01H7FDSCXX2PJ4GNPE2V5GR0EC', now(), now(), '01H4Z1V2X401KBR5B102N9ZJ4T') 
ON CONFLICT DO NOTHING;
        
-- (-2147483629) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXY4SY6F6JEAX9XDQ19', 'ChatSystemAdmin', true, now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GGVXWS1AFVBW5ADTX2PWDX2Y', '01H7FDSCXY4SY6F6JEAX9XDQ19', now(), now(), '-2147483629'),
    ('01GGVXWS16ERRH43KEYM2NB4EH', '01H7FDSCXY4SY6F6JEAX9XDQ19', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYJ0ERV79T17NN8YZE', 'ChatSystemAdmin', true, now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYV5WP1EHP89TZ1PC4', '01H7FDSCXYJ0ERV79T17NN8YZE', '01H7FDSCXY4SY6F6JEAX9XDQ19', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYV5WP1EHP89TZ1PC4', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXYZEMWV781WZQMY8V1', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483629', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXYZEMWV781WZQMY8V1', '01H7FDSCXYJ0ERV79T17NN8YZE', now(), now(), '01GFMNHQ1WHGRC8AW6K913AM3G') 
ON CONFLICT DO NOTHING;
        
-- (2147483629) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXYXDYGBK8GHQ1A4HEX', 'ChatSystemAdmin', true, now(), now(), '2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4Z60EJW02AT7VC103BT4T8X', '01H7FDSCXYXDYGBK8GHQ1A4HEX', now(), now(), '2147483629'),
    ('01H4Z60EJX00XQC0YS00T4Z3X7', '01H7FDSCXYXDYGBK8GHQ1A4HEX', now(), now(), '2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYJFXTQNW0QTHVKW6A', 'ChatSystemAdmin', true, now(), now(), '2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYS5HAQD952Q06JTAT', '01H7FDSCXYJFXTQNW0QTHVKW6A', '01H7FDSCXYXDYGBK8GHQ1A4HEX', now(), now(), '2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXYS5HAQD952Q06JTAT', '01H4Z60EJT02KKDKGN01S81X0H', now(), now(), '2147483629')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXYYBDNP1EMCNC0DMGD', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483629', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXYYBDNP1EMCNC0DMGD', '01H7FDSCXYJFXTQNW0QTHVKW6A', now(), now(), '01H4Z60EJT02KKDKGN01S81X0H') 
ON CONFLICT DO NOTHING;
        
-- (-2147483628) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXYF6BNR9CF1244JKEN', 'ChatSystemAdmin', true, now(), now(), '-2147483628')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GS7PYBJDPQ321F4YG48EBZQM', '01H7FDSCXYF6BNR9CF1244JKEN', now(), now(), '-2147483628'),
    ('01GS7PXY7JGS75MCN3Q5M6RD22', '01H7FDSCXYF6BNR9CF1244JKEN', now(), now(), '-2147483628')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZF4X4AZXCBFX1Y3S4', 'ChatSystemAdmin', true, now(), now(), '-2147483628')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZ47MJW95BWKKE2ZS8', '01H7FDSCXZF4X4AZXCBFX1Y3S4', '01H7FDSCXYF6BNR9CF1244JKEN', now(), now(), '-2147483628')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZ47MJW95BWKKE2ZS8', '01GRB92TYDRPXMVAHPYXTSHFT9', now(), now(), '-2147483628')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXZMFRWT7JQNHDR723X', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483628', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXZMFRWT7JQNHDR723X', '01H7FDSCXZF4X4AZXCBFX1Y3S4', now(), now(), '01GRB92TYDRPXMVAHPYXTSHFT9') 
ON CONFLICT DO NOTHING;
        
-- (-2147483627) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXZZ39WDEXC8KG7P07E', 'ChatSystemAdmin', true, now(), now(), '-2147483627')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GTBCS2HHH088XKZR0JBRF9EX', '01H7FDSCXZZ39WDEXC8KG7P07E', now(), now(), '-2147483627'),
    ('01GTBCS8V0DZTAD33H5MVM79TS', '01H7FDSCXZZ39WDEXC8KG7P07E', now(), now(), '-2147483627')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZNJAKANJCKX40QGQK', 'ChatSystemAdmin', true, now(), now(), '-2147483627')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZHRPHDC53T05RG8S1', '01H7FDSCXZNJAKANJCKX40QGQK', '01H7FDSCXZZ39WDEXC8KG7P07E', now(), now(), '-2147483627')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZHRPHDC53T05RG8S1', '01GTBAS9GYFGQ6C39VF75QNV6Q', now(), now(), '-2147483627')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXZ67FJ79DVJR5DPAS5', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483627', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXZ67FJ79DVJR5DPAS5', '01H7FDSCXZNJAKANJCKX40QGQK', now(), now(), '01GTBAS9GYFGQ6C39VF75QNV6Q') 
ON CONFLICT DO NOTHING;
        
-- (-2147483626) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXZHH20734V9GA9SD41', 'ChatSystemAdmin', true, now(), now(), '-2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GX2R5EFJNJSWJH11TWSJB11S', '01H7FDSCXZHH20734V9GA9SD41', now(), now(), '-2147483626'),
    ('01GX2R589M5P99AQ96C3W2XQDR', '01H7FDSCXZHH20734V9GA9SD41', now(), now(), '-2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZDNXSPB3QHHH9B545', 'ChatSystemAdmin', true, now(), now(), '-2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZZV1CWWWBM1HCPZAJ', '01H7FDSCXZDNXSPB3QHHH9B545', '01H7FDSCXZHH20734V9GA9SD41', now(), now(), '-2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZZV1CWWWBM1HCPZAJ', '01GX2R4MH7FXFKDH26JKMV91Q5', now(), now(), '-2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCXZFDN4W67ZDRBC851Z', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483626', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCXZFDN4W67ZDRBC851Z', '01H7FDSCXZDNXSPB3QHHH9B545', now(), now(), '01GX2R4MH7FXFKDH26JKMV91Q5') 
ON CONFLICT DO NOTHING;
        
-- (2147483626) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCXZWCQKQ6W1HCAZ39X0', 'ChatSystemAdmin', true, now(), now(), '2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4Z64D5K03WV68BF02WMVNG9', '01H7FDSCXZWCQKQ6W1HCAZ39X0', now(), now(), '2147483626'),
    ('01H4Z64D5K01FZGSDX034E7CEC', '01H7FDSCXZWCQKQ6W1HCAZ39X0', now(), now(), '2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZRA4FXS9YHEYQDD2M', 'ChatSystemAdmin', true, now(), now(), '2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZ5PCDCABX2XTYB7CP', '01H7FDSCXZRA4FXS9YHEYQDD2M', '01H7FDSCXZWCQKQ6W1HCAZ39X0', now(), now(), '2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCXZ5PCDCABX2XTYB7CP', '01H4Z64D5H01N0X163029ZS285', now(), now(), '2147483626')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY0W1SKFXH9YBR720DQ', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483626', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY0W1SKFXH9YBR720DQ', '01H7FDSCXZRA4FXS9YHEYQDD2M', now(), now(), '01H4Z64D5H01N0X163029ZS285') 
ON CONFLICT DO NOTHING;
        
-- (-2147483625) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY1HQFPRPJ7WJAAFYVP', 'ChatSystemAdmin', true, now(), now(), '-2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GY9JKBZ22VBTSKJYCTYM7H5X', '01H7FDSCY1HQFPRPJ7WJAAFYVP', now(), now(), '-2147483625'),
    ('01GY9JKVG04F9MBKV8HHR8N341', '01H7FDSCY1HQFPRPJ7WJAAFYVP', now(), now(), '-2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1GM6R3ZBR5035AKAK', 'ChatSystemAdmin', true, now(), now(), '-2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1DAZXQ2P4NS5GRSJQ', '01H7FDSCY1GM6R3ZBR5035AKAK', '01H7FDSCY1HQFPRPJ7WJAAFYVP', now(), now(), '-2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1DAZXQ2P4NS5GRSJQ', '01GY9KYKRW4M16YPB0V03JRTE6', now(), now(), '-2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY1ZGJQ7WS6WX3XGD4S', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483625', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY1ZGJQ7WS6WX3XGD4S', '01H7FDSCY1GM6R3ZBR5035AKAK', now(), now(), '01GY9KYKRW4M16YPB0V03JRTE6') 
ON CONFLICT DO NOTHING;
        
-- (2147483625) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY1K7D6GAAQTM86RVS2', 'ChatSystemAdmin', true, now(), now(), '2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4Z66YAE03NCHM1601HK3291', '01H7FDSCY1K7D6GAAQTM86RVS2', now(), now(), '2147483625'),
    ('01H4Z66YAD00SWFBK603CVJTNZ', '01H7FDSCY1K7D6GAAQTM86RVS2', now(), now(), '2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY11K7DE1XHJSPJFX1E', 'ChatSystemAdmin', true, now(), now(), '2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1SPY4G677ZB1GMWHE', '01H7FDSCY11K7DE1XHJSPJFX1E', '01H7FDSCY1K7D6GAAQTM86RVS2', now(), now(), '2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1SPY4G677ZB1GMWHE', '01H4Z66YAB00CDJ6A701S32BNF', now(), now(), '2147483625')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY1XSK2KMVDQR2H7A7J', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483625', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY1XSK2KMVDQR2H7A7J', '01H7FDSCY11K7DE1XHJSPJFX1E', now(), now(), '01H4Z66YAB00CDJ6A701S32BNF') 
ON CONFLICT DO NOTHING;
        
-- (-2147483624) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY1E8QCP6HX75VNEENJ', 'ChatSystemAdmin', true, now(), now(), '-2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GZJDXF68ZMJMWP00VG62SDNV', '01H7FDSCY1E8QCP6HX75VNEENJ', now(), now(), '-2147483624'),
    ('01GZJDXMXN6M5ZCXDYZTJ574XD', '01H7FDSCY1E8QCP6HX75VNEENJ', now(), now(), '-2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY185D7EAZR6GW2V0GG', 'ChatSystemAdmin', true, now(), now(), '-2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1J11D2F0G8932MACX', '01H7FDSCY185D7EAZR6GW2V0GG', '01H7FDSCY1E8QCP6HX75VNEENJ', now(), now(), '-2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1J11D2F0G8932MACX', '01GZJDV9GCHMYD8MKGT7CV052W', now(), now(), '-2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY1MRE9A48FJ2ADVQJS', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483624', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY1MRE9A48FJ2ADVQJS', '01H7FDSCY185D7EAZR6GW2V0GG', now(), now(), '01GZJDV9GCHMYD8MKGT7CV052W') 
ON CONFLICT DO NOTHING;
        
-- (2147483624) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY1ABGK4RXEZTX7K0BQ', 'ChatSystemAdmin', true, now(), now(), '2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H4Z91ZB301SGH6BX01QKB5P8', '01H7FDSCY1ABGK4RXEZTX7K0BQ', now(), now(), '2147483624'),
    ('01H4Z91ZB400EVSR4D011N7S8C', '01H7FDSCY1ABGK4RXEZTX7K0BQ', now(), now(), '2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1MJTVST7WGB9SM3ZM', 'ChatSystemAdmin', true, now(), now(), '2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1FWNSD4YAB9RW1P5R', '01H7FDSCY1MJTVST7WGB9SM3ZM', '01H7FDSCY1ABGK4RXEZTX7K0BQ', now(), now(), '2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY1FWNSD4YAB9RW1P5R', '01H4Z91ZB103RK3FRB03H5Z152', now(), now(), '2147483624')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY192WDFXTDQ4S3Z0Q9', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2147483624', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY192WDFXTDQ4S3Z0Q9', '01H7FDSCY1MJTVST7WGB9SM3ZM', now(), now(), '01H4Z91ZB103RK3FRB03H5Z152') 
ON CONFLICT DO NOTHING;
        
-- (-2147483623) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY1TRVPS2SJKJQMSWQS', 'ChatSystemAdmin', true, now(), now(), '-2147483623')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GZZJE7NSE2QENXZJ7SBX2X5D', '01H7FDSCY1TRVPS2SJKJQMSWQS', now(), now(), '-2147483623'),
    ('01GZZJECW1HF0M6142G2HYW71R', '01H7FDSCY1TRVPS2SJKJQMSWQS', now(), now(), '-2147483623')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2CED1AY9B1PJ4EMSS', 'ChatSystemAdmin', true, now(), now(), '-2147483623')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2CW2W6RYXDH7W6X66', '01H7FDSCY2CED1AY9B1PJ4EMSS', '01H7FDSCY1TRVPS2SJKJQMSWQS', now(), now(), '-2147483623')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2CW2W6RYXDH7W6X66', '01GZZJE0A5AMPCCF41GA830NEK', now(), now(), '-2147483623')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY2Z7V05M55QB5KEZ4X', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483623', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY2Z7V05M55QB5KEZ4X', '01H7FDSCY2CED1AY9B1PJ4EMSS', now(), now(), '01GZZJE0A5AMPCCF41GA830NEK') 
ON CONFLICT DO NOTHING;
        
-- (-2147483622) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY2JPFFW5SEY2XPVA1S', 'ChatSystemAdmin', true, now(), now(), '-2147483622')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H1BAJ3T993PB3JNVMT33G040', '01H7FDSCY2JPFFW5SEY2XPVA1S', now(), now(), '-2147483622'),
    ('01H1BAHW9TJBTH2M1QYY0HMSCK', '01H7FDSCY2JPFFW5SEY2XPVA1S', now(), now(), '-2147483622')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY22SBQ5MA3F18QT8M7', 'ChatSystemAdmin', true, now(), now(), '-2147483622')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2P28RAY5VYWVCKJP3', '01H7FDSCY22SBQ5MA3F18QT8M7', '01H7FDSCY2JPFFW5SEY2XPVA1S', now(), now(), '-2147483622')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2P28RAY5VYWVCKJP3', '01H1BAHG1YXMMXEVSTS7NBV9CH', now(), now(), '-2147483622')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY22JCYVGTZAEQ2ESYN', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483622', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY22JCYVGTZAEQ2ESYN', '01H7FDSCY22SBQ5MA3F18QT8M7', now(), now(), '01H1BAHG1YXMMXEVSTS7NBV9CH') 
ON CONFLICT DO NOTHING;
        
-- (100013) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY2JZZT79QD9C5663B1', 'ChatSystemAdmin', true, now(), now(), '100013')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H5P628QR005SF1VY01JXGGSJ', '01H7FDSCY2JZZT79QD9C5663B1', now(), now(), '100013'),
    ('01H5P628QR004N91HE002K6KGK', '01H7FDSCY2JZZT79QD9C5663B1', now(), now(), '100013')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY257VNBMSQYX2F8634', 'ChatSystemAdmin', true, now(), now(), '100013')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2PEF1T9DNVEKD5Q9Y', '01H7FDSCY257VNBMSQYX2F8634', '01H7FDSCY2JZZT79QD9C5663B1', now(), now(), '100013')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2PEF1T9DNVEKD5Q9Y', '01H5P628QM02RN623302TH23AQ', now(), now(), '100013')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY2PG9Y54KRC1JZZ5XA', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '100013', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY2PG9Y54KRC1JZZ5XA', '01H7FDSCY257VNBMSQYX2F8634', now(), now(), '01H5P628QM02RN623302TH23AQ') 
ON CONFLICT DO NOTHING;
        
-- (100012) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY2EXKBH641MFGY4WYM', 'ChatSystemAdmin', true, now(), now(), '100012')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01H1XGQ07FCAX49RPMFT9QQMJP', '01H7FDSCY2EXKBH641MFGY4WYM', now(), now(), '100012'),
    ('01H1XGPQQQ8H0CA8ABJTBQFVCT', '01H7FDSCY2EXKBH641MFGY4WYM', now(), now(), '100012')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2XRYP4RH4K8HYD1XB', 'ChatSystemAdmin', true, now(), now(), '100012')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY225ABB3WQ4EC6R2P2', '01H7FDSCY2XRYP4RH4K8HYD1XB', '01H7FDSCY2EXKBH641MFGY4WYM', now(), now(), '100012')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY225ABB3WQ4EC6R2P2', '01H1XGH5WK9Z11DM21V75JYBSE', now(), now(), '100012')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY20CGNV304V9M7Y9GS', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '100012', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY20CGNV304V9M7Y9GS', '01H7FDSCY2XRYP4RH4K8HYD1XB', now(), now(), '01H1XGH5WK9Z11DM21V75JYBSE') 
ON CONFLICT DO NOTHING;
        
-- (100000) --
INSERT INTO role (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
    ('01H7FDSCY2J9VMTFVE15GGY4TH', 'ChatSystemAdmin', true, now(), now(), '100000')
ON CONFLICT DO NOTHING;
        
INSERT INTO permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
    ('01GMMSRDYPDX9H0XMVFM3ADQJA', '01H7FDSCY2J9VMTFVE15GGY4TH', now(), now(), '100000'),
    ('01GMMSRDYPDX9H0XMVGJMSVRPS', '01H7FDSCY2J9VMTFVE15GGY4TH', now(), now(), '100000')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.user_group (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2NH0MZJMBYSW25W18', 'ChatSystemAdmin', true, now(), now(), '100000')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2PNKQ6A8KHPCEMEGT', '01H7FDSCY2NH0MZJMBYSW25W18', '01H7FDSCY2J9VMTFVE15GGY4TH', now(), now(), '100000')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.granted_role_access_path (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01H7FDSCY2PNKQ6A8KHPCEMEGT', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000')
ON CONFLICT DO NOTHING;
        
INSERT INTO public.users (user_id, "country", "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name, encrypted_user_id_by_password, deactivated_at, username, login_email, user_role)
VALUES
    ('01H7FDSCY2RPT8XZ1P53D0X91T', 'COUNTRY_JP', 'ChatSystemAdmin', '', NULL, 'chat_system_admin+user@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '100000', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, true, NULL, NULL, NULL, NULL, NULL, NULL, 'staff')
ON CONFLICT DO NOTHING;
        
INSERT INTO user_group_member(user_id,  user_group_id, created_at, updated_at, resource_path) 
VALUES 
    ('01H7FDSCY2RPT8XZ1P53D0X91T', '01H7FDSCY2NH0MZJMBYSW25W18', now(), now(), '911FLMNMYA6SKTTT44HWE2E100') 
ON CONFLICT DO NOTHING;
        