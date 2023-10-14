--- Add permission ---
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GTZYYSBEEVBQXREAYRR37PA1', 'payment.product.read', now(), now(), '-2147483627'),
  ('01GTZYYSBEEVBQXREAYRR37PA2', 'payment.product.write', now(), now(), '-2147483627'),
  ('01GTZYYSBEEVBQXREAYRR37PB1', 'payment.product.read', now(), now(), '-2147483628'),
  ('01GTZYYSBEEVBQXREAYRR37PB2', 'payment.product.write', now(), now(), '-2147483628'),
  ('01GTZYYSBEEVBQXREAYRR37QA1', 'payment.product.read', now(), now(), '-2147483629'),
  ('01GTZYYSBEEVBQXREAYRR37QA2', 'payment.product.write', now(), now(), '-2147483629'),
  ('01GTZYYSBEEVBQXREAYRR37QB1', 'payment.product.read', now(), now(), '-2147483630'),
  ('01GTZYYSBEEVBQXREAYRR37QB2', 'payment.product.write', now(), now(), '-2147483630'),
  ('01GTZYYSBEEVBQXREAYRR37QC1', 'payment.product.read', now(), now(), '-2147483631'),
  ('01GTZYYSBEEVBQXREAYRR37QC2', 'payment.product.write', now(), now(), '-2147483631'),
  ('01GTZYYSBEEVBQXREAYRR37QD1', 'payment.product.read', now(), now(), '-2147483632'),
  ('01GTZYYSBEEVBQXREAYRR37QD2', 'payment.product.write', now(), now(), '-2147483632'),
  ('01GTZYYSBEEVBQXREAYRR37QE1', 'payment.product.read', now(), now(), '-2147483633'),
  ('01GTZYYSBEEVBQXREAYRR37QE2', 'payment.product.write', now(), now(), '-2147483633'),
  ('01GTZYYSBEEVBQXREAYRR37QF1', 'payment.product.read', now(), now(), '-2147483634'),
  ('01GTZYYSBEEVBQXREAYRR37QF2', 'payment.product.write', now(), now(), '-2147483634'),
  ('01GTZYYSBEEVBQXREAYRR37QG1', 'payment.product.read', now(), now(), '-2147483635'),
  ('01GTZYYSBEEVBQXREAYRR37QG2', 'payment.product.write', now(), now(), '-2147483635'),
  ('01GTZYYSBEEVBQXREAYRR37QH1', 'payment.product.read', now(), now(), '-2147483637'),
  ('01GTZYYSBEEVBQXREAYRR37QH2', 'payment.product.write', now(), now(), '-2147483637'),
  ('01GTZYYSBEEVBQXREAYRR37QI1', 'payment.product.read', now(), now(), '-2147483638'),
  ('01GTZYYSBEEVBQXREAYRR37QI2', 'payment.product.write', now(), now(), '-2147483638'),
  ('01GTZYYSBEEVBQXREAYRR37QJ1', 'payment.product.read', now(), now(), '-2147483639'),
  ('01GTZYYSBEEVBQXREAYRR37QJ2', 'payment.product.write', now(), now(), '-2147483639'),
  ('01GTZYYSBEEVBQXREAYRR37QK1', 'payment.product.read', now(), now(), '-2147483640'),
  ('01GTZYYSBEEVBQXREAYRR37QK2', 'payment.product.write', now(), now(), '-2147483640'),
  ('01GTZYYSBEEVBQXREAYRR37QL1', 'payment.product.read', now(), now(), '-2147483641'),
  ('01GTZYYSBEEVBQXREAYRR37QL2', 'payment.product.write', now(), now(), '-2147483641'),
  ('01GTZYYSBEEVBQXREAYRR37QM1', 'payment.product.read', now(), now(), '-2147483642'),
  ('01GTZYYSBEEVBQXREAYRR37QM2', 'payment.product.write', now(), now(), '-2147483642'),
  ('01GTZYYSBEEVBQXREAYRR37QN1', 'payment.product.read', now(), now(), '-2147483643'),
  ('01GTZYYSBEEVBQXREAYRR37QN2', 'payment.product.write', now(), now(), '-2147483643'),
  ('01GTZYYSBEEVBQXREAYRR37QO1', 'payment.product.read', now(), now(), '-2147483644'),
  ('01GTZYYSBEEVBQXREAYRR37QO2', 'payment.product.write', now(), now(), '-2147483644'),
  ('01GTZYYSBEEVBQXREAYRR37QP1', 'payment.product.read', now(), now(), '-2147483645'),
  ('01GTZYYSBEEVBQXREAYRR37QP2', 'payment.product.write', now(), now(), '-2147483645'),
  ('01GTZYYSBEEVBQXREAYRR37QR1', 'payment.product.read', now(), now(), '-2147483646'),
  ('01GTZYYSBEEVBQXREAYRR37QR2', 'payment.product.write', now(), now(), '-2147483646'),
  ('01GTZYYSBEEVBQXREAYRR37QS1', 'payment.product.read', now(), now(), '-2147483647'),
  ('01GTZYYSBEEVBQXREAYRR37QS2', 'payment.product.write', now(), now(), '-2147483647'),
  ('01GTZYYSBEEVBQXREAYRR37QT1', 'payment.product.read', now(), now(), '-2147483648'),
  ('01GTZYYSBEEVBQXREAYRR37QT2', 'payment.product.write', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Add role ---   
INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01GTZYYSEPYTSDW16YPQR872A1', 'PaymentScheduleJob', true, now(), now(), '-2147483627'),
  ('01GTZYYSEPYTSDW16YPQR872A2', 'PaymentScheduleJob', true, now(), now(), '-2147483628'),
  ('01GTZYYSEPYTSDW16YPQR873A1', 'PaymentScheduleJob', true, now(), now(), '-2147483629'),
  ('01GTZYYSEPYTSDW16YPQR873A2', 'PaymentScheduleJob', true, now(), now(), '-2147483630'),
  ('01GTZYYSEPYTSDW16YPQR873A3', 'PaymentScheduleJob', true, now(), now(), '-2147483631'),
  ('01GTZYYSEPYTSDW16YPQR873A4', 'PaymentScheduleJob', true, now(), now(), '-2147483632'),
  ('01GTZYYSEPYTSDW16YPQR873A5', 'PaymentScheduleJob', true, now(), now(), '-2147483633'),
  ('01GTZYYSEPYTSDW16YPQR873A6', 'PaymentScheduleJob', true, now(), now(), '-2147483634'),
  ('01GTZYYSEPYTSDW16YPQR873A7', 'PaymentScheduleJob', true, now(), now(), '-2147483635'),
  ('01GTZYYSEPYTSDW16YPQR873A8', 'PaymentScheduleJob', true, now(), now(), '-2147483637'),
  ('01GTZYYSEPYTSDW16YPQR873A9', 'PaymentScheduleJob', true, now(), now(), '-2147483638'),
  ('01GTZYYSEPYTSDW16YPQR873B1', 'PaymentScheduleJob', true, now(), now(), '-2147483639'),
  ('01GTZYYSEPYTSDW16YPQR873B2', 'PaymentScheduleJob', true, now(), now(), '-2147483640'),
  ('01GTZYYSEPYTSDW16YPQR873B3', 'PaymentScheduleJob', true, now(), now(), '-2147483641'),
  ('01GTZYYSEPYTSDW16YPQR873B4', 'PaymentScheduleJob', true, now(), now(), '-2147483642'),
  ('01GTZYYSEPYTSDW16YPQR873B5', 'PaymentScheduleJob', true, now(), now(), '-2147483643'),
  ('01GTZYYSEPYTSDW16YPQR873B6', 'PaymentScheduleJob', true, now(), now(), '-2147483644'),
  ('01GTZYYSEPYTSDW16YPQR873B7', 'PaymentScheduleJob', true, now(), now(), '-2147483645'),
  ('01GTZYYSEPYTSDW16YPQR873B8', 'PaymentScheduleJob', true, now(), now(), '-2147483646'),
  ('01GTZYYSEPYTSDW16YPQR873B9', 'PaymentScheduleJob', true, now(), now(), '-2147483647'),
  ('01GTZYYSEPYTSDW16YPQR873C1', 'PaymentScheduleJob', true, now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Add permission_role ---   
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GTZYYSBEEVBQXREAYRR37PA1', '01GTZYYSEPYTSDW16YPQR872A1', now(), now(), '-2147483627'),
  ('01GTZYYSBEEVBQXREAYRR37PA2', '01GTZYYSEPYTSDW16YPQR872A1', now(), now(), '-2147483627'),
  ('01GTZYYSBEEVBQXREAYRR37PB1', '01GTZYYSEPYTSDW16YPQR872A2', now(), now(), '-2147483628'),
  ('01GTZYYSBEEVBQXREAYRR37PB2', '01GTZYYSEPYTSDW16YPQR872A2', now(), now(), '-2147483628'),
  ('01GTZYYSBEEVBQXREAYRR37QA1', '01GTZYYSEPYTSDW16YPQR873A1', now(), now(), '-2147483629'),
  ('01GTZYYSBEEVBQXREAYRR37QA2', '01GTZYYSEPYTSDW16YPQR873A1', now(), now(), '-2147483629'),
  ('01GTZYYSBEEVBQXREAYRR37QB1', '01GTZYYSEPYTSDW16YPQR873A2', now(), now(), '-2147483630'),
  ('01GTZYYSBEEVBQXREAYRR37QB2', '01GTZYYSEPYTSDW16YPQR873A2', now(), now(), '-2147483630'),
  ('01GTZYYSBEEVBQXREAYRR37QC1', '01GTZYYSEPYTSDW16YPQR873A3', now(), now(), '-2147483631'),
  ('01GTZYYSBEEVBQXREAYRR37QC2', '01GTZYYSEPYTSDW16YPQR873A3', now(), now(), '-2147483631'),
  ('01GTZYYSBEEVBQXREAYRR37QD1', '01GTZYYSEPYTSDW16YPQR873A4', now(), now(), '-2147483632'),
  ('01GTZYYSBEEVBQXREAYRR37QD2', '01GTZYYSEPYTSDW16YPQR873A4', now(), now(), '-2147483632'),
  ('01GTZYYSBEEVBQXREAYRR37QE1', '01GTZYYSEPYTSDW16YPQR873A5', now(), now(), '-2147483633'),
  ('01GTZYYSBEEVBQXREAYRR37QE2', '01GTZYYSEPYTSDW16YPQR873A5', now(), now(), '-2147483633'),
  ('01GTZYYSBEEVBQXREAYRR37QF1', '01GTZYYSEPYTSDW16YPQR873A6', now(), now(), '-2147483634'),
  ('01GTZYYSBEEVBQXREAYRR37QF2', '01GTZYYSEPYTSDW16YPQR873A6', now(), now(), '-2147483634'),
  ('01GTZYYSBEEVBQXREAYRR37QG1', '01GTZYYSEPYTSDW16YPQR873A7', now(), now(), '-2147483635'),
  ('01GTZYYSBEEVBQXREAYRR37QG2', '01GTZYYSEPYTSDW16YPQR873A7', now(), now(), '-2147483635'),
  ('01GTZYYSBEEVBQXREAYRR37QH1', '01GTZYYSEPYTSDW16YPQR873A8', now(), now(), '-2147483637'),
  ('01GTZYYSBEEVBQXREAYRR37QH2', '01GTZYYSEPYTSDW16YPQR873A8', now(), now(), '-2147483637'),
  ('01GTZYYSBEEVBQXREAYRR37QI1', '01GTZYYSEPYTSDW16YPQR873A9', now(), now(), '-2147483638'),
  ('01GTZYYSBEEVBQXREAYRR37QI2', '01GTZYYSEPYTSDW16YPQR873A9', now(), now(), '-2147483638'),
  ('01GTZYYSBEEVBQXREAYRR37QJ1', '01GTZYYSEPYTSDW16YPQR873B1', now(), now(), '-2147483639'),
  ('01GTZYYSBEEVBQXREAYRR37QJ2', '01GTZYYSEPYTSDW16YPQR873B1', now(), now(), '-2147483639'),
  ('01GTZYYSBEEVBQXREAYRR37QK1', '01GTZYYSEPYTSDW16YPQR873B2', now(), now(), '-2147483640'),
  ('01GTZYYSBEEVBQXREAYRR37QK2', '01GTZYYSEPYTSDW16YPQR873B2', now(), now(), '-2147483640'),
  ('01GTZYYSBEEVBQXREAYRR37QL1', '01GTZYYSEPYTSDW16YPQR873B3', now(), now(), '-2147483641'),
  ('01GTZYYSBEEVBQXREAYRR37QL2', '01GTZYYSEPYTSDW16YPQR873B3', now(), now(), '-2147483641'),
  ('01GTZYYSBEEVBQXREAYRR37QM1', '01GTZYYSEPYTSDW16YPQR873B4', now(), now(), '-2147483642'),
  ('01GTZYYSBEEVBQXREAYRR37QM2', '01GTZYYSEPYTSDW16YPQR873B4', now(), now(), '-2147483642'),
  ('01GTZYYSBEEVBQXREAYRR37QN1', '01GTZYYSEPYTSDW16YPQR873B5', now(), now(), '-2147483643'),
  ('01GTZYYSBEEVBQXREAYRR37QN2', '01GTZYYSEPYTSDW16YPQR873B5', now(), now(), '-2147483643'),
  ('01GTZYYSBEEVBQXREAYRR37QO1', '01GTZYYSEPYTSDW16YPQR873B6', now(), now(), '-2147483644'),
  ('01GTZYYSBEEVBQXREAYRR37QO2', '01GTZYYSEPYTSDW16YPQR873B6', now(), now(), '-2147483644'),
  ('01GTZYYSBEEVBQXREAYRR37QP1', '01GTZYYSEPYTSDW16YPQR873B7', now(), now(), '-2147483645'),
  ('01GTZYYSBEEVBQXREAYRR37QP2', '01GTZYYSEPYTSDW16YPQR873B7', now(), now(), '-2147483645'),
  ('01GTZYYSBEEVBQXREAYRR37QR1', '01GTZYYSEPYTSDW16YPQR873B8', now(), now(), '-2147483646'),
  ('01GTZYYSBEEVBQXREAYRR37QR2', '01GTZYYSEPYTSDW16YPQR873B8', now(), now(), '-2147483646'),
  ('01GTZYYSBEEVBQXREAYRR37QS1', '01GTZYYSEPYTSDW16YPQR873B9', now(), now(), '-2147483647'),
  ('01GTZYYSBEEVBQXREAYRR37QS2', '01GTZYYSEPYTSDW16YPQR873B9', now(), now(), '-2147483647'),
  ('01GTZYYSBEEVBQXREAYRR37QT1', '01GTZYYSEPYTSDW16YPQR873C1', now(), now(), '-2147483648'),
  ('01GTZYYSBEEVBQXREAYRR37QT2', '01GTZYYSEPYTSDW16YPQR873C1', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Add User Group --- 
INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GV00DNYKP9CDSD30MMQF23A1', 'PaymentScheduleJob', true, now(), now(), '-2147483627'),
  ('01GV00DNYKP9CDSD30MMQF23A2', 'PaymentScheduleJob', true, now(), now(), '-2147483628'),
  ('01GV00DNYKP9CDSD30MMQF24A1', 'PaymentScheduleJob', true, now(), now(), '-2147483629'),
  ('01GV00DNYKP9CDSD30MMQF24A2', 'PaymentScheduleJob', true, now(), now(), '-2147483630'),
  ('01GV00DNYKP9CDSD30MMQF24A3', 'PaymentScheduleJob', true, now(), now(), '-2147483631'),
  ('01GV00DNYKP9CDSD30MMQF24A4', 'PaymentScheduleJob', true, now(), now(), '-2147483632'),
  ('01GV00DNYKP9CDSD30MMQF24A5', 'PaymentScheduleJob', true, now(), now(), '-2147483633'),
  ('01GV00DNYKP9CDSD30MMQF24A6', 'PaymentScheduleJob', true, now(), now(), '-2147483634'),
  ('01GV00DNYKP9CDSD30MMQF24A7', 'PaymentScheduleJob', true, now(), now(), '-2147483635'),
  ('01GV00DNYKP9CDSD30MMQF24A8', 'PaymentScheduleJob', true, now(), now(), '-2147483637'),
  ('01GV00DNYKP9CDSD30MMQF24A9', 'PaymentScheduleJob', true, now(), now(), '-2147483638'),
  ('01GV00DNYKP9CDSD30MMQF24B1', 'PaymentScheduleJob', true, now(), now(), '-2147483639'),
  ('01GV00DNYKP9CDSD30MMQF24B2', 'PaymentScheduleJob', true, now(), now(), '-2147483640'),
  ('01GV00DNYKP9CDSD30MMQF24B3', 'PaymentScheduleJob', true, now(), now(), '-2147483641'),
  ('01GV00DNYKP9CDSD30MMQF24B4', 'PaymentScheduleJob', true, now(), now(), '-2147483642'),
  ('01GV00DNYKP9CDSD30MMQF24B5', 'PaymentScheduleJob', true, now(), now(), '-2147483643'),
  ('01GV00DNYKP9CDSD30MMQF24B6', 'PaymentScheduleJob', true, now(), now(), '-2147483644'),
  ('01GV00DNYKP9CDSD30MMQF24B7', 'PaymentScheduleJob', true, now(), now(), '-2147483645'),
  ('01GV00DNYKP9CDSD30MMQF24B8', 'PaymentScheduleJob', true, now(), now(), '-2147483646'),
  ('01GV00DNYKP9CDSD30MMQF24B9', 'PaymentScheduleJob', true, now(), now(), '-2147483647'),
  ('01GV00DNYKP9CDSD30MMQF24C1', 'PaymentScheduleJob', true, now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Grant role to User group ---
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GFWRE5WXWA24F8TVDJHTZ3B1', '01GV00DNYKP9CDSD30MMQF23A1', '01GTZYYSEPYTSDW16YPQR872A1', now(), now(), '-2147483627'),
  ('01GFWRE5WXWA24F8TVDJHTZ3B2', '01GV00DNYKP9CDSD30MMQF23A2', '01GTZYYSEPYTSDW16YPQR872A2', now(), now(), '-2147483628'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B1', '01GV00DNYKP9CDSD30MMQF24A1', '01GTZYYSEPYTSDW16YPQR873A1', now(), now(), '-2147483629'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B2', '01GV00DNYKP9CDSD30MMQF24A2', '01GTZYYSEPYTSDW16YPQR873A2', now(), now(), '-2147483630'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B3', '01GV00DNYKP9CDSD30MMQF24A3', '01GTZYYSEPYTSDW16YPQR873A3', now(), now(), '-2147483631'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B4', '01GV00DNYKP9CDSD30MMQF24A4', '01GTZYYSEPYTSDW16YPQR873A4', now(), now(), '-2147483632'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B5', '01GV00DNYKP9CDSD30MMQF24A5', '01GTZYYSEPYTSDW16YPQR873A5', now(), now(), '-2147483633'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B6', '01GV00DNYKP9CDSD30MMQF24A6', '01GTZYYSEPYTSDW16YPQR873A6', now(), now(), '-2147483634'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B7', '01GV00DNYKP9CDSD30MMQF24A7', '01GTZYYSEPYTSDW16YPQR873A7', now(), now(), '-2147483635'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B8', '01GV00DNYKP9CDSD30MMQF24A8', '01GTZYYSEPYTSDW16YPQR873A8', now(), now(), '-2147483637'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B9', '01GV00DNYKP9CDSD30MMQF24A9', '01GTZYYSEPYTSDW16YPQR873A9', now(), now(), '-2147483638'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C1', '01GV00DNYKP9CDSD30MMQF24B1', '01GTZYYSEPYTSDW16YPQR873B1', now(), now(), '-2147483639'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C2', '01GV00DNYKP9CDSD30MMQF24B2', '01GTZYYSEPYTSDW16YPQR873B2', now(), now(), '-2147483640'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C3', '01GV00DNYKP9CDSD30MMQF24B3', '01GTZYYSEPYTSDW16YPQR873B3', now(), now(), '-2147483641'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C4', '01GV00DNYKP9CDSD30MMQF24B4', '01GTZYYSEPYTSDW16YPQR873B4', now(), now(), '-2147483642'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C5', '01GV00DNYKP9CDSD30MMQF24B5', '01GTZYYSEPYTSDW16YPQR873B5', now(), now(), '-2147483643'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C6', '01GV00DNYKP9CDSD30MMQF24B6', '01GTZYYSEPYTSDW16YPQR873B6', now(), now(), '-2147483644'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C7', '01GV00DNYKP9CDSD30MMQF24B7', '01GTZYYSEPYTSDW16YPQR873B7', now(), now(), '-2147483645'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C8', '01GV00DNYKP9CDSD30MMQF24B8', '01GTZYYSEPYTSDW16YPQR873B8', now(), now(), '-2147483646'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C9', '01GV00DNYKP9CDSD30MMQF24B9', '01GTZYYSEPYTSDW16YPQR873B9', now(), now(), '-2147483647'),
  ('01GFWRE5WXWA24F8TVDJHTZ4D1', '01GV00DNYKP9CDSD30MMQF24C1', '01GTZYYSEPYTSDW16YPQR873C1', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Grant location to a role --- 
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GFWRE5WXWA24F8TVDJHTZ3B1', '01GTBAS9GYFGQ6C39VF75QNV6Q', now(), now(), '-2147483627'),
  ('01GFWRE5WXWA24F8TVDJHTZ3B2', '01GRB92TYDRPXMVAHPYXTSHFT9', now(), now(), '-2147483628'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B1', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B2', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B3', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B6', '01FR4M51XJY9E77GSN4QZ1Q8N5', now(), now(), '-2147483634'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B7', '01FR4M51XJY9E77GSN4QZ1Q8N4', now(), now(), '-2147483635'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B8', '01FR4M51XJY9E77GSN4QZ1Q8N3', now(), now(), '-2147483637'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B9', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C1', '01FR4M51XJY9E77GSN4QZ1Q8N1', now(), now(), '-2147483639'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C2', '01FR4M51XJY9E77GSN4QZ1Q9N9', now(), now(), '-2147483640'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C3', '01FR4M51XJY9E77GSN4QZ1Q9N8', now(), now(), '-2147483641'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C4', '01FR4M51XJY9E77GSN4QZ1Q9N7', now(), now(), '-2147483642'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C5', '01FR4M51XJY9E77GSN4QZ1Q9N6', now(), now(), '-2147483643'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C6', '01FR4M51XJY9E77GSN4QZ1Q9N5', now(), now(), '-2147483644'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C7', '01FR4M51XJY9E77GSN4QZ1Q9N4', now(), now(), '-2147483645'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C8', '01FR4M51XJY9E77GSN4QZ1Q9N3', now(), now(), '-2147483646'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C9', '01FR4M51XJY9E77GSN4QZ1Q9N2', now(), now(), '-2147483647'),
  ('01GFWRE5WXWA24F8TVDJHTZ4D1', '01FR4M51XJY9E77GSN4QZ1Q9N1', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Upsert granted_permission ---
INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF23A1')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF23A2')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A1')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A2')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A3')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A4')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A5')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A6')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A7')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A8')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24A9')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B1')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B2')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B3')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B4')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B5')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B6')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B7')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B8')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24B9')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission 
  (user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path) 
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF24C1')
ON CONFLICT ON CONSTRAINT granted_permission__pk 
DO UPDATE SET user_group_name = excluded.user_group_name;
