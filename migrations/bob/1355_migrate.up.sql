-- Step introduction per organization:
-- 1. Delete redundant permission role
-- Redundant permission is added in file: 1338_migrate.up.sql
-- Ignore organization: -2147483629, -2147483630, -2147483631, -2147483632, -2147483633 (no redundant)
-- + NotificationScheduleJob - communication.notification.read
-- + NotificationScheduleJob - communication.notification.write

-- 2. Delete duplicate permission (communication.notification.read, communication.notification.write)
-- Ignore organization: -2147483629, -2147483630, -2147483631, -2147483632, -2147483633 (no duplicate)

-- 3. Add permission needed for role NotificationScheduleJob: 
-- Ignore organization: -2147483629, -2147483630, -2147483631, -2147483632, -2147483633 (others permission haven't added yet)
-- + communication.notification.read
-- + communication.notification.write
-- + master.location.read (support migrate location data)
-- + master.course.read (support migrate course data)
-- + master.class.read (support migrate class data)
-- + user.student.read (read data to get recipient)
-- + user.user.read (read data to assign name for recipient)

-- 4. Add record for notification_internal_user table
-- + Added in file: 1353_migrate.up.sql


-- Doing:
-- resource_path -2147483629 --
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB1', true, now(), now(), NULL, '-2147483629');

-- resource_path -2147483630 --
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB2', true, now(), now(), NULL, '-2147483630');

-- resource_path -2147483631 --
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB3', true, now(), now(), NULL, '-2147483631');

-- resource_path -2147483632 --
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB4', true, now(), now(), NULL, '-2147483632');

-- resource_path -2147483633 --
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB5', true, now(), now(), NULL, '-2147483633');

-- resource_path -2147483634 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ16E2FM3860RAWS388Y61' AND role_id = '01GFWQC67NF55GYEXBNRR12AD6' AND resource_path = '-2147483634';
DELETE FROM permission_role WHERE permission_id = '01GFWQ16E2FM3860RAWS388Y62' AND role_id = '01GFWQC67NF55GYEXBNRR12AD6' AND resource_path = '-2147483634';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ16E2FM3860RAWS388Y61' AND resource_path = '-2147483634';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ16E2FM3860RAWS388Y62' AND resource_path = '-2147483634';
DELETE FROM permission WHERE permission_id = '01GFWQ16E2FM3860RAWS388Y61' AND resource_path = '-2147483634';
DELETE FROM permission WHERE permission_id = '01GFWQ16E2FM3860RAWS388Y62' AND resource_path = '-2147483634';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHQPYSTT9X424GAWBZGFAE1', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GDHQPYSTT9X424GAWBZGFAE2', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GFAGZJ00GZY7RNZRVZBMM7VR', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GFAGCPC67XAN2F0K93JB2SRD', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GCGH2Q9F381YRDP4E4WZ45X0', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GG9EWWVNBHGDM01WQERA8MX0', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB6', true, now(), now(), NULL, '-2147483634');

-- resource_path -2147483635 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ1EV1VW1T105SP9M4D4G1' AND role_id = '01GFWQC67NF55GYEXBNRR12AD7' AND resource_path = '-2147483635';
DELETE FROM permission_role WHERE permission_id = '01GFWQ1EV1VW1T105SP9M4D4G2' AND role_id = '01GFWQC67NF55GYEXBNRR12AD7' AND resource_path = '-2147483635';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ1EV1VW1T105SP9M4D4G1' AND resource_path = '-2147483635';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ1EV1VW1T105SP9M4D4G2' AND resource_path = '-2147483635';
DELETE FROM permission WHERE permission_id = '01GFWQ1EV1VW1T105SP9M4D4G1' AND resource_path = '-2147483635';
DELETE FROM permission WHERE permission_id = '01GFWQ1EV1VW1T105SP9M4D4G2' AND resource_path = '-2147483635';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHQSX3792NFE3475J01CMQ1', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GDHQSX3792NFE3475J01CMQ2', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GFAGZJ00GZY7RNZRW0C8BW0Z', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GFAGCPC67XAN2F0K93JXRWDN', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01G8T46TPSRMAKTMRPR7Q93ZFK', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GCGH2Q9F381YRDP4E4XY9ZF0', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GG9EWWVNBHGDM01WQERA8MX2', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB7', true, now(), now(), NULL, '-2147483635');

-- resource_path -2147483637 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ249YWDVDFPC0WA5NYGD1' AND role_id = '01GFWQC67NF55GYEXBNRR12AD8' AND resource_path = '-2147483637';
DELETE FROM permission_role WHERE permission_id = '01GFWQ249YWDVDFPC0WA5NYGD2' AND role_id = '01GFWQC67NF55GYEXBNRR12AD8' AND resource_path = '-2147483637';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ249YWDVDFPC0WA5NYGD1' AND resource_path = '-2147483637';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ249YWDVDFPC0WA5NYGD2' AND resource_path = '-2147483637';
DELETE FROM permission WHERE permission_id = '01GFWQ249YWDVDFPC0WA5NYGD1' AND resource_path = '-2147483637';
DELETE FROM permission WHERE permission_id = '01GFWQ249YWDVDFPC0WA5NYGD2' AND resource_path = '-2147483637';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHQWEWMWR4NKR8TY1BS30Q1', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GDHQWEWMWR4NKR8TY1BS30Q2', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GFAGZJ00GZY7RNZRW2GG00BV', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GFAGCPC67XAN2F0K9504X7QF', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01G6A5YH5MS80ZEPS9CSNW8KPP', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GCGH2Q9F381YRDP4EA7ENAT0', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GG9EWWVNBHGDM01WQERA8MX4', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB8', true, now(), now(), NULL, '-2147483637');

-- resource_path -2147483638 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ2Y3RXKHXJTWEC0NYTZA1' AND role_id = '01GFWQC67NF55GYEXBNRR12AD9' AND resource_path = '-2147483638';
DELETE FROM permission_role WHERE permission_id = '01GFWQ2Y3RXKHXJTWEC0NYTZA2' AND role_id = '01GFWQC67NF55GYEXBNRR12AD9' AND resource_path = '-2147483638';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ2Y3RXKHXJTWEC0NYTZA1' AND resource_path = '-2147483638';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ2Y3RXKHXJTWEC0NYTZA2' AND resource_path = '-2147483638';
DELETE FROM permission WHERE permission_id = '01GFWQ2Y3RXKHXJTWEC0NYTZA1' AND resource_path = '-2147483638';
DELETE FROM permission WHERE permission_id = '01GFWQ2Y3RXKHXJTWEC0NYTZA2' AND resource_path = '-2147483638';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHR22CJWMHTYMCTPNM96EK1', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GDHR22CJWMHTYMCTPNM96EK2', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GFAGZJ00GZY7RNZRW3C344TY', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GFAGCPC67XAN2F0K97P67K4W', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01G6A69MMBG2A02KRDQE1F2QPT', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GCGH2Q9F381YRDP4EBDQZDW0', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GG9EWWVNBHGDM01WQERA8MX6', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZB9', true, now(), now(), NULL, '-2147483638');

-- resource_path -2147483639 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ35K0EPQV5ASA0MT7F4B1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE1' AND resource_path = '-2147483639';
DELETE FROM permission_role WHERE permission_id = '01GFWQ35K0EPQV5ASA0MT7F4B2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE1' AND resource_path = '-2147483639';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ35K0EPQV5ASA0MT7F4B1' AND resource_path = '-2147483639';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ35K0EPQV5ASA0MT7F4B2' AND resource_path = '-2147483639';
DELETE FROM permission WHERE permission_id = '01GFWQ35K0EPQV5ASA0MT7F4B1' AND resource_path = '-2147483639';
DELETE FROM permission WHERE permission_id = '01GFWQ35K0EPQV5ASA0MT7F4B2' AND resource_path = '-2147483639';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHR6CTY0XZ7JBA2F81515A1', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GDHR6CTY0XZ7JBA2F81515A2', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GFAGZJ00GZY7RNZRW3FEE8WH', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GFAGCPC67XAN2F0K99W57RMF', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01G1GQ13MVRCRHVP79GDPZTCY1', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GCGH2Q9F381YRDP4EF3831Q0', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GG9EWWVNBHGDM01WQERA8MX8', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC1', true, now(), now(), NULL, '-2147483639');

-- resource_path -2147483640 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ3H8SXWDNZ23MA7GXWGK1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE2' AND resource_path = '-2147483640';
DELETE FROM permission_role WHERE permission_id = '01GFWQ3H8SXWDNZ23MA7GXWGK2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE2' AND resource_path = '-2147483640';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ3H8SXWDNZ23MA7GXWGK1' AND resource_path = '-2147483640';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ3H8SXWDNZ23MA7GXWGK2' AND resource_path = '-2147483640';
DELETE FROM permission WHERE permission_id = '01GFWQ3H8SXWDNZ23MA7GXWGK1' AND resource_path = '-2147483640';
DELETE FROM permission WHERE permission_id = '01GFWQ3H8SXWDNZ23MA7GXWGK2' AND resource_path = '-2147483640';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHRB8HRNDN65W6K16JKWGT1', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GDHRB8HRNDN65W6K16JKWGT2', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GFAGZJ00GZY7RNZRW4XF5J2A', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GFAGCPC67XAN2F0K9C1Z18SN', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ9', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GCGH2Q9F381YRDP4EFBBBKP0', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GDMTJ3NB0TKRWS7ZRMHFTHH8', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC2', true, now(), now(), NULL, '-2147483640');

-- resource_path -2147483641 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ3RM0M7BSEA3QD2HHDXK1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE3' AND resource_path = '-2147483641';
DELETE FROM permission_role WHERE permission_id = '01GFWQ3RM0M7BSEA3QD2HHDXK2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE3' AND resource_path = '-2147483641';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ3RM0M7BSEA3QD2HHDXK1' AND resource_path = '-2147483641';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ3RM0M7BSEA3QD2HHDXK2' AND resource_path = '-2147483641';
DELETE FROM permission WHERE permission_id = '01GFWQ3RM0M7BSEA3QD2HHDXK1' AND resource_path = '-2147483641';
DELETE FROM permission WHERE permission_id = '01GFWQ3RM0M7BSEA3QD2HHDXK2' AND resource_path = '-2147483641';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHRHZZSFQ2WQ2KG5J5DNFE1', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GDHRHZZSFQ2WQ2KG5J5DNFE2', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GFAGZJ00GZY7RNZRW8321H4G', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GFAGCPC67XAN2F0K9E5DE3R2', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ8', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GCGH2Q9F381YRDP4EFD8HWN0', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GDMTJ3NB0TKRWS7ZRMHFTHH6', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC3', true, now(), now(), NULL, '-2147483641');

-- resource_path -2147483642 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ44RA8N1D2RJ573WRD8S1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE4' AND resource_path = '-2147483642';
DELETE FROM permission_role WHERE permission_id = '01GFWQ44RA8N1D2RJ573WRD8S2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE4' AND resource_path = '-2147483642';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ44RA8N1D2RJ573WRD8S1' AND resource_path = '-2147483642';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ44RA8N1D2RJ573WRD8S2' AND resource_path = '-2147483642';
DELETE FROM permission WHERE permission_id = '01GFWQ44RA8N1D2RJ573WRD8S1' AND resource_path = '-2147483642';
DELETE FROM permission WHERE permission_id = '01GFWQ44RA8N1D2RJ573WRD8S2' AND resource_path = '-2147483642';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHRWA9YH1VAGDP01CFYEPP1', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GDHRWA9YH1VAGDP01CFYEPP2', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GFAGZJ00GZY7RNZRWA0BBDSB', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GFAGCPC67XAN2F0K9H63H18D', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ7', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GCGH2Q9F381YRDP4EG135FB0', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GDMTJ3NB0TKRWS7ZRMHFTHH4', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC4', true, now(), now(), NULL, '-2147483642');

-- resource_path -2147483643 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ4C9YDDHZKK3SM7KZD5N1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE5' AND resource_path = '-2147483643';
DELETE FROM permission_role WHERE permission_id = '01GFWQ4C9YDDHZKK3SM7KZD5N2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE5' AND resource_path = '-2147483643';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ4C9YDDHZKK3SM7KZD5N1' AND resource_path = '-2147483643';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ4C9YDDHZKK3SM7KZD5N2' AND resource_path = '-2147483643';
DELETE FROM permission WHERE permission_id = '01GFWQ4C9YDDHZKK3SM7KZD5N1' AND resource_path = '-2147483643';
DELETE FROM permission WHERE permission_id = '01GFWQ4C9YDDHZKK3SM7KZD5N2' AND resource_path = '-2147483643';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHYB4ZNZHR46P2MW08A7TG1', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GDHYB4ZNZHR46P2MW08A7TG2', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GFAGZJ00GZY7RNZRWA5TZDXC', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GFAGCPC67XAN2F0K9J4DRQHJ', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ6', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GCGH2Q9F381YRDP4EHRQ26E0', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GDMTJ3NB0TKRWS7ZRMHFTHH0', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC5', true, now(), now(), NULL, '-2147483643');

-- resource_path -2147483644 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ4MFSE87NS1HJXT44EFY1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE6' AND resource_path = '-2147483644';
DELETE FROM permission_role WHERE permission_id = '01GFWQ4MFSE87NS1HJXT44EFY2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE6' AND resource_path = '-2147483644';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ4MFSE87NS1HJXT44EFY1' AND resource_path = '-2147483644';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ4MFSE87NS1HJXT44EFY2' AND resource_path = '-2147483644';
DELETE FROM permission WHERE permission_id = '01GFWQ4MFSE87NS1HJXT44EFY1' AND resource_path = '-2147483644';
DELETE FROM permission WHERE permission_id = '01GFWQ4MFSE87NS1HJXT44EFY2' AND resource_path = '-2147483644';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHXJ1527M7FNVBD9DQKD9Y1', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GDHXJ1527M7FNVBD9DQKD9Y2', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GFAGZJ00GZY7RNZRWBTNGPZ0', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GFAGCPC67XAN2F0K9KKM5Q81', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ5', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GCGH2Q9F381YRDP4EJFDXJH0', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GDMTJ3NB0TKRWS7ZRMHFTHH1', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC6', true, now(), now(), NULL, '-2147483644');

-- resource_path -2147483645 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ4V5SK6PMFX15TD725A11' AND role_id = '01GFWQC67NF55GYEXBNRR12AE7' AND resource_path = '-2147483645';
DELETE FROM permission_role WHERE permission_id = '01GFWQ4V5SK6PMFX15TD725A12' AND role_id = '01GFWQC67NF55GYEXBNRR12AE7' AND resource_path = '-2147483645';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ4V5SK6PMFX15TD725A11' AND resource_path = '-2147483645';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ4V5SK6PMFX15TD725A12' AND resource_path = '-2147483645';
DELETE FROM permission WHERE permission_id = '01GFWQ4V5SK6PMFX15TD725A11' AND resource_path = '-2147483645';
DELETE FROM permission WHERE permission_id = '01GFWQ4V5SK6PMFX15TD725A12' AND resource_path = '-2147483645';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHXN8C0YFAA4TE432SFCXQ1', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GDHXN8C0YFAA4TE432SFCXQ2', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GFAGZJ00GZY7RNZRWDZ5H0SK', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GFAGCPC67XAN2F0K9PKZQ8QV', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ4', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GCGH2Q9F381YRDP4EKNM7DN0', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GG9EWWVPFT4HTY1E7RF9F4B0', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC7', true, now(), now(), NULL, '-2147483645');

-- resource_path -2147483646 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ51TTQRD1TPTF97MJ3X91' AND role_id = '01GFWQC67NF55GYEXBNRR12AE8' AND resource_path = '-2147483646';
DELETE FROM permission_role WHERE permission_id = '01GFWQ51TTQRD1TPTF97MJ3X92' AND role_id = '01GFWQC67NF55GYEXBNRR12AE8' AND resource_path = '-2147483646';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ51TTQRD1TPTF97MJ3X91' AND resource_path = '-2147483646';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ51TTQRD1TPTF97MJ3X92' AND resource_path = '-2147483646';
DELETE FROM permission WHERE permission_id = '01GFWQ51TTQRD1TPTF97MJ3X91' AND resource_path = '-2147483646';
DELETE FROM permission WHERE permission_id = '01GFWQ51TTQRD1TPTF97MJ3X92' AND resource_path = '-2147483646';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHXVJV3JKFR221AJ5HDAX01', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GDHXVJV3JKFR221AJ5HDAX02', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GFAGZJ00GZY7RNZRWE1FVWR7', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GFAGCPC67XAN2F0K9PVJBM8P', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ3', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GCGH2Q9F381YRDP4ENC72ZR0', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GG9EWWVPFT4HTY1E7RF9F4B2', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC8', true, now(), now(), NULL, '-2147483646');

-- resource_path -2147483647 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ58413YRG57J21CRKCXR1' AND role_id = '01GFWQC67NF55GYEXBNRR12AE9' AND resource_path = '-2147483647';
DELETE FROM permission_role WHERE permission_id = '01GFWQ58413YRG57J21CRKCXR2' AND role_id = '01GFWQC67NF55GYEXBNRR12AE9' AND resource_path = '-2147483647';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ58413YRG57J21CRKCXR1' AND resource_path = '-2147483647';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ58413YRG57J21CRKCXR2' AND resource_path = '-2147483647';
DELETE FROM permission WHERE permission_id = '01GFWQ58413YRG57J21CRKCXR1' AND resource_path = '-2147483647';
DELETE FROM permission WHERE permission_id = '01GFWQ58413YRG57J21CRKCXR2' AND resource_path = '-2147483647';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHXX98F2V5HQ39C7ACXJDG1', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GDHXX98F2V5HQ39C7ACXJDG2', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GFAGZJ00GZY7RNZRWH5N68FX', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GFAGCPC67XAN2F0K9STCP257', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ2', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GCGH2Q9F381YRDP4ERVG0JC0', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GG9EWWVPFT4HTY1E7RF9F4B4', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZC9', true, now(), now(), NULL, '-2147483647');

-- resource_path -2147483648 --
DELETE FROM permission_role WHERE permission_id = '01GFWQ5EY9SCS2AZ8GTDDWSMP1' AND role_id = '01GFWQC67NF55GYEXBNRR12AF1' AND resource_path = '-2147483648';
DELETE FROM permission_role WHERE permission_id = '01GFWQ5EY9SCS2AZ8GTDDWSMP2' AND role_id = '01GFWQC67NF55GYEXBNRR12AF1' AND resource_path = '-2147483648';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ5EY9SCS2AZ8GTDDWSMP1' AND resource_path = '-2147483648';
DELETE FROM granted_permission WHERE permission_id = '01GFWQ5EY9SCS2AZ8GTDDWSMP2' AND resource_path = '-2147483648';
DELETE FROM permission WHERE permission_id = '01GFWQ5EY9SCS2AZ8GTDDWSMP1' AND resource_path = '-2147483648';
DELETE FROM permission WHERE permission_id = '01GFWQ5EY9SCS2AZ8GTDDWSMP2' AND resource_path = '-2147483648';

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GDHY1VFR1XAH795YVFX1D1P1', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01GDHY1VFR1XAH795YVFX1D1P2', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01GFAGZJ00GZY7RNZRWKJKX174', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01GFAGCPC67XAN2F0K9T3YQ78T', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01G1GQ13MVRCRHVP79GDPZTCZ1', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01GCGH2Q9F381YRDP4ETZ937G0', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01GG9EWWVPFT4HTY1E7RF9F4B6', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01GGP4SS7CNK30HXKC9CMKAZD1', true, now(), now(), NULL, '-2147483648');

--- Upsert granted_permission ---
INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S1')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S2')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S3')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S4')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S5')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S6')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S7')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S8')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S9')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A1')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A2')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A3')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A4')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A5')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A6')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A7')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A8')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8A9')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8B1')
ON CONFLICT ON CONSTRAINT granted_permission__uniq 
DO UPDATE SET user_group_name = excluded.user_group_name;
