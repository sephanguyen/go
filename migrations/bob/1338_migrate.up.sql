--- Add permission ---
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GFWPT9WFAZBSR84N14HRTGN1', 'communication.notification.read', now(), now(), '-2147483629'),
  ('01GFWPT9WFAZBSR84N14HRTGN2', 'communication.notification.write', now(), now(), '-2147483629'),
  ('01GFWPV54P3P49ZPT3BB069X21', 'communication.notification.read', now(), now(), '-2147483630'),
  ('01GFWPV54P3P49ZPT3BB069X22', 'communication.notification.write', now(), now(), '-2147483630'),
  ('01GFWPW8R69G0MXSTWSPACHJ31', 'communication.notification.read', now(), now(), '-2147483631'),
  ('01GFWPW8R69G0MXSTWSPACHJ32', 'communication.notification.write', now(), now(), '-2147483631'),
  ('01GFWPX3ANP0C9KP51TC50WVF1', 'communication.notification.read', now(), now(), '-2147483632'),
  ('01GFWPX3ANP0C9KP51TC50WVF2', 'communication.notification.write', now(), now(), '-2147483632'),
  ('01GFWQ0SRJDM8WPRFB4BSW9S41', 'communication.notification.read', now(), now(), '-2147483633'),
  ('01GFWQ0SRJDM8WPRFB4BSW9S42', 'communication.notification.write', now(), now(), '-2147483633'),
  ('01GFWQ16E2FM3860RAWS388Y61', 'communication.notification.read', now(), now(), '-2147483634'),
  ('01GFWQ16E2FM3860RAWS388Y62', 'communication.notification.write', now(), now(), '-2147483634'),
  ('01GFWQ1EV1VW1T105SP9M4D4G1', 'communication.notification.read', now(), now(), '-2147483635'),
  ('01GFWQ1EV1VW1T105SP9M4D4G2', 'communication.notification.write', now(), now(), '-2147483635'),
  ('01GFWQ249YWDVDFPC0WA5NYGD1', 'communication.notification.read', now(), now(), '-2147483637'),
  ('01GFWQ249YWDVDFPC0WA5NYGD2', 'communication.notification.write', now(), now(), '-2147483637'),
  ('01GFWQ2Y3RXKHXJTWEC0NYTZA1', 'communication.notification.read', now(), now(), '-2147483638'),
  ('01GFWQ2Y3RXKHXJTWEC0NYTZA2', 'communication.notification.write', now(), now(), '-2147483638'),
  ('01GFWQ35K0EPQV5ASA0MT7F4B1', 'communication.notification.read', now(), now(), '-2147483639'),
  ('01GFWQ35K0EPQV5ASA0MT7F4B2', 'communication.notification.write', now(), now(), '-2147483639'),
  ('01GFWQ3H8SXWDNZ23MA7GXWGK1', 'communication.notification.read', now(), now(), '-2147483640'),
  ('01GFWQ3H8SXWDNZ23MA7GXWGK2', 'communication.notification.write', now(), now(), '-2147483640'),
  ('01GFWQ3RM0M7BSEA3QD2HHDXK1', 'communication.notification.read', now(), now(), '-2147483641'),
  ('01GFWQ3RM0M7BSEA3QD2HHDXK2', 'communication.notification.write', now(), now(), '-2147483641'),
  ('01GFWQ44RA8N1D2RJ573WRD8S1', 'communication.notification.read', now(), now(), '-2147483642'),
  ('01GFWQ44RA8N1D2RJ573WRD8S2', 'communication.notification.write', now(), now(), '-2147483642'),
  ('01GFWQ4C9YDDHZKK3SM7KZD5N1', 'communication.notification.read', now(), now(), '-2147483643'),
  ('01GFWQ4C9YDDHZKK3SM7KZD5N2', 'communication.notification.write', now(), now(), '-2147483643'),
  ('01GFWQ4MFSE87NS1HJXT44EFY1', 'communication.notification.read', now(), now(), '-2147483644'),
  ('01GFWQ4MFSE87NS1HJXT44EFY2', 'communication.notification.write', now(), now(), '-2147483644'),
  ('01GFWQ4V5SK6PMFX15TD725A11', 'communication.notification.read', now(), now(), '-2147483645'),
  ('01GFWQ4V5SK6PMFX15TD725A12', 'communication.notification.write', now(), now(), '-2147483645'),
  ('01GFWQ51TTQRD1TPTF97MJ3X91', 'communication.notification.read', now(), now(), '-2147483646'),
  ('01GFWQ51TTQRD1TPTF97MJ3X92', 'communication.notification.write', now(), now(), '-2147483646'),
  ('01GFWQ58413YRG57J21CRKCXR1', 'communication.notification.read', now(), now(), '-2147483647'),
  ('01GFWQ58413YRG57J21CRKCXR2', 'communication.notification.write', now(), now(), '-2147483647'),
  ('01GFWQ5EY9SCS2AZ8GTDDWSMP1', 'communication.notification.read', now(), now(), '-2147483648'),
  ('01GFWQ5EY9SCS2AZ8GTDDWSMP2', 'communication.notification.write', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Add role ---   
INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01GFWQC67NF55GYEXBNRR12AD1', 'NotificationScheduleJob', true, now(), now(), '-2147483629'),
  ('01GFWQC67NF55GYEXBNRR12AD2', 'NotificationScheduleJob', true, now(), now(), '-2147483630'),
  ('01GFWQC67NF55GYEXBNRR12AD3', 'NotificationScheduleJob', true, now(), now(), '-2147483631'),
  ('01GFWQC67NF55GYEXBNRR12AD4', 'NotificationScheduleJob', true, now(), now(), '-2147483632'),
  ('01GFWQC67NF55GYEXBNRR12AD5', 'NotificationScheduleJob', true, now(), now(), '-2147483633'),
  ('01GFWQC67NF55GYEXBNRR12AD6', 'NotificationScheduleJob', true, now(), now(), '-2147483634'),
  ('01GFWQC67NF55GYEXBNRR12AD7', 'NotificationScheduleJob', true, now(), now(), '-2147483635'),
  ('01GFWQC67NF55GYEXBNRR12AD8', 'NotificationScheduleJob', true, now(), now(), '-2147483637'),
  ('01GFWQC67NF55GYEXBNRR12AD9', 'NotificationScheduleJob', true, now(), now(), '-2147483638'),
  ('01GFWQC67NF55GYEXBNRR12AE1', 'NotificationScheduleJob', true, now(), now(), '-2147483639'),
  ('01GFWQC67NF55GYEXBNRR12AE2', 'NotificationScheduleJob', true, now(), now(), '-2147483640'),
  ('01GFWQC67NF55GYEXBNRR12AE3', 'NotificationScheduleJob', true, now(), now(), '-2147483641'),
  ('01GFWQC67NF55GYEXBNRR12AE4', 'NotificationScheduleJob', true, now(), now(), '-2147483642'),
  ('01GFWQC67NF55GYEXBNRR12AE5', 'NotificationScheduleJob', true, now(), now(), '-2147483643'),
  ('01GFWQC67NF55GYEXBNRR12AE6', 'NotificationScheduleJob', true, now(), now(), '-2147483644'),
  ('01GFWQC67NF55GYEXBNRR12AE7', 'NotificationScheduleJob', true, now(), now(), '-2147483645'),
  ('01GFWQC67NF55GYEXBNRR12AE8', 'NotificationScheduleJob', true, now(), now(), '-2147483646'),
  ('01GFWQC67NF55GYEXBNRR12AE9', 'NotificationScheduleJob', true, now(), now(), '-2147483647'),
  ('01GFWQC67NF55GYEXBNRR12AF1', 'NotificationScheduleJob', true, now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Add permission_role ---   
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GFWPT9WFAZBSR84N14HRTGN1', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629'),
  ('01GFWPT9WFAZBSR84N14HRTGN2', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629'),
  ('01GFWPV54P3P49ZPT3BB069X21', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630'),
  ('01GFWPV54P3P49ZPT3BB069X22', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630'),
  ('01GFWPW8R69G0MXSTWSPACHJ31', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631'),
  ('01GFWPW8R69G0MXSTWSPACHJ32', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631'),
  ('01GFWPX3ANP0C9KP51TC50WVF1', '01GFWQC67NF55GYEXBNRR12AD4', now(), now(), '-2147483632'),
  ('01GFWPX3ANP0C9KP51TC50WVF2', '01GFWQC67NF55GYEXBNRR12AD4', now(), now(), '-2147483632'),
  ('01GFWQ0SRJDM8WPRFB4BSW9S41', '01GFWQC67NF55GYEXBNRR12AD5', now(), now(), '-2147483633'),
  ('01GFWQ0SRJDM8WPRFB4BSW9S42', '01GFWQC67NF55GYEXBNRR12AD5', now(), now(), '-2147483633'),
  ('01GFWQ16E2FM3860RAWS388Y61', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GFWQ16E2FM3860RAWS388Y62', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GFWQ1EV1VW1T105SP9M4D4G1', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GFWQ1EV1VW1T105SP9M4D4G2', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GFWQ249YWDVDFPC0WA5NYGD1', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GFWQ249YWDVDFPC0WA5NYGD2', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GFWQ2Y3RXKHXJTWEC0NYTZA1', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GFWQ2Y3RXKHXJTWEC0NYTZA2', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GFWQ35K0EPQV5ASA0MT7F4B1', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GFWQ35K0EPQV5ASA0MT7F4B2', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GFWQ3H8SXWDNZ23MA7GXWGK1', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GFWQ3H8SXWDNZ23MA7GXWGK2', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GFWQ3RM0M7BSEA3QD2HHDXK1', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GFWQ3RM0M7BSEA3QD2HHDXK2', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GFWQ44RA8N1D2RJ573WRD8S1', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GFWQ44RA8N1D2RJ573WRD8S2', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GFWQ4C9YDDHZKK3SM7KZD5N1', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GFWQ4C9YDDHZKK3SM7KZD5N2', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GFWQ4MFSE87NS1HJXT44EFY1', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GFWQ4MFSE87NS1HJXT44EFY2', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GFWQ4V5SK6PMFX15TD725A11', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GFWQ4V5SK6PMFX15TD725A12', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GFWQ51TTQRD1TPTF97MJ3X91', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GFWQ51TTQRD1TPTF97MJ3X92', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GFWQ58413YRG57J21CRKCXR1', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GFWQ58413YRG57J21CRKCXR2', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GFWQ5EY9SCS2AZ8GTDDWSMP1', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648'),
  ('01GFWQ5EY9SCS2AZ8GTDDWSMP2', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Add User Group --- 
INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GFWR7AGF3KQ6SHDTMDH7S8S1', 'NotificationScheduleJob', true, now(), now(), '-2147483629'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S2', 'NotificationScheduleJob', true, now(), now(), '-2147483630'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S3', 'NotificationScheduleJob', true, now(), now(), '-2147483631'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S4', 'NotificationScheduleJob', true, now(), now(), '-2147483632'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S5', 'NotificationScheduleJob', true, now(), now(), '-2147483633'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S6', 'NotificationScheduleJob', true, now(), now(), '-2147483634'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S7', 'NotificationScheduleJob', true, now(), now(), '-2147483635'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S8', 'NotificationScheduleJob', true, now(), now(), '-2147483637'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8S9', 'NotificationScheduleJob', true, now(), now(), '-2147483638'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A1', 'NotificationScheduleJob', true, now(), now(), '-2147483639'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A2', 'NotificationScheduleJob', true, now(), now(), '-2147483640'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A3', 'NotificationScheduleJob', true, now(), now(), '-2147483641'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A4', 'NotificationScheduleJob', true, now(), now(), '-2147483642'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A5', 'NotificationScheduleJob', true, now(), now(), '-2147483643'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A6', 'NotificationScheduleJob', true, now(), now(), '-2147483644'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A7', 'NotificationScheduleJob', true, now(), now(), '-2147483645'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A8', 'NotificationScheduleJob', true, now(), now(), '-2147483646'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8A9', 'NotificationScheduleJob', true, now(), now(), '-2147483647'),
  ('01GFWR7AGF3KQ6SHDTMDH7S8B1', 'NotificationScheduleJob', true, now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Grant role to User group ---
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GFWRE5WXWA24F8TVDJHTZ4B1', '01GFWR7AGF3KQ6SHDTMDH7S8S1', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B2', '01GFWR7AGF3KQ6SHDTMDH7S8S2', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B3', '01GFWR7AGF3KQ6SHDTMDH7S8S3', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B4', '01GFWR7AGF3KQ6SHDTMDH7S8S4', '01GFWQC67NF55GYEXBNRR12AD4', now(), now(), '-2147483632'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B5', '01GFWR7AGF3KQ6SHDTMDH7S8S5', '01GFWQC67NF55GYEXBNRR12AD5', now(), now(), '-2147483633'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B6', '01GFWR7AGF3KQ6SHDTMDH7S8S6', '01GFWQC67NF55GYEXBNRR12AD6', now(), now(), '-2147483634'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B7', '01GFWR7AGF3KQ6SHDTMDH7S8S7', '01GFWQC67NF55GYEXBNRR12AD7', now(), now(), '-2147483635'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B8', '01GFWR7AGF3KQ6SHDTMDH7S8S8', '01GFWQC67NF55GYEXBNRR12AD8', now(), now(), '-2147483637'),
  ('01GFWRE5WXWA24F8TVDJHTZ4B9', '01GFWR7AGF3KQ6SHDTMDH7S8S9', '01GFWQC67NF55GYEXBNRR12AD9', now(), now(), '-2147483638'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C1', '01GFWR7AGF3KQ6SHDTMDH7S8A1', '01GFWQC67NF55GYEXBNRR12AE1', now(), now(), '-2147483639'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C2', '01GFWR7AGF3KQ6SHDTMDH7S8A2', '01GFWQC67NF55GYEXBNRR12AE2', now(), now(), '-2147483640'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C3', '01GFWR7AGF3KQ6SHDTMDH7S8A3', '01GFWQC67NF55GYEXBNRR12AE3', now(), now(), '-2147483641'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C4', '01GFWR7AGF3KQ6SHDTMDH7S8A4', '01GFWQC67NF55GYEXBNRR12AE4', now(), now(), '-2147483642'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C5', '01GFWR7AGF3KQ6SHDTMDH7S8A5', '01GFWQC67NF55GYEXBNRR12AE5', now(), now(), '-2147483643'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C6', '01GFWR7AGF3KQ6SHDTMDH7S8A6', '01GFWQC67NF55GYEXBNRR12AE6', now(), now(), '-2147483644'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C7', '01GFWR7AGF3KQ6SHDTMDH7S8A7', '01GFWQC67NF55GYEXBNRR12AE7', now(), now(), '-2147483645'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C8', '01GFWR7AGF3KQ6SHDTMDH7S8A8', '01GFWQC67NF55GYEXBNRR12AE8', now(), now(), '-2147483646'),
  ('01GFWRE5WXWA24F8TVDJHTZ4C9', '01GFWR7AGF3KQ6SHDTMDH7S8A9', '01GFWQC67NF55GYEXBNRR12AE9', now(), now(), '-2147483647'),
  ('01GFWRE5WXWA24F8TVDJHTZ4D1', '01GFWR7AGF3KQ6SHDTMDH7S8B1', '01GFWQC67NF55GYEXBNRR12AF1', now(), now(), '-2147483648')
  ON CONFLICT DO NOTHING;

--- Grant location to a role --- 
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
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
