------------ Managara Base -------------
--- Location ---
INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01GFMMFRXC6SKTTT44HWR3BRY8','org','Managara Base', '-2147483630', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('01GFMMFRXC6SKTTT44HWR3BRY8', 'Managara Base','01GFMMFRXC6SKTTT44HWR3BRY8',NULL, NULL, NULL, '-2147483630', now(), now(),'01GFMMFRXC6SKTTT44HWR3BRY8') ON CONFLICT DO NOTHING;

--- Organization ---
INSERT INTO organizations (organization_id, tenant_id,               name,       resource_path, domain_name, logo_url, country,      created_at, updated_at, deleted_at)
VALUES                    ('-2147483630',   'withus-managara-base-mrkvu', 'Managara Base', '-2147483630', 'managara-base', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant-managara-base-logo.png',     'COUNTRY_JP', now(),      now(),      null      ) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483630, 'student-coach-e1e95', 'withus-managara-base-mrkvu') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id,   name,        country,      city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES              (-2147483630, 'Managara Base', 'COUNTRY_JP', 1,       1,           null,  false,            now(),      now(),      false,    null,         null,       '-2147483630') ON CONFLICT DO NOTHING;

--- Dynamic form ---
INSERT
	INTO
	public.partner_form_configs (form_config_id,
	partner_id,
	feature_name,
	created_at,
	updated_at,
	deleted_at,
	form_config_data,
	resource_path)
VALUES
		('01GFMMFRXDZYGZMDAPBS2NGB1C',
-2147483630,
'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT',
now(),
now(),
NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_required": true, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_remark", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REMARK"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Homework Submission", "ja": "課題", "vi": "Homework Submission"}, "fallback_language": "ja"}}, "field_id": "homework_submission_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework Status", "ja": "提出状況", "vi": "Homework Status"}, "fallback_language": "ja"}}, "field_id": "homework_submission_status", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "label": {"i18n": {"translations": {"en": "Completed", "ja": "完了", "vi": "Completed"}, "fallback_language": "ja"}}}, {"key": "INCOMPLETE", "label": {"i18n": {"translations": {"en": "Incomplete", "ja": "未完了", "vi": "Incomplete"}, "fallback_language": "ja"}}}], "optionLabelKey": "label"}, "component_config": {"type": "AUTOCOMPLETE"}}], "section_id": "homework_submission_section_id", "section_name": "homework_submission"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Lesson", "ja": "授業", "vi": "Lesson"}, "fallback_language": "ja"}}, "field_id": "lesson_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_view_study_plan_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "LINK_VIEW_STUDY_PLAN"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "追加教材", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "追加課題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "lesson_section_id", "section_name": "lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_section_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "remarks_section_id", "section_name": "remarks"}]}',
'-2147483630')
	ON
CONFLICT ON
CONSTRAINT partner_form_configs_pk DO NOTHING;

--- Role <> User Group ---
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', 'master.location.read', now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;
INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01GFMMFRXDZHGVC3YWC7J668F1', 'Teacher', false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F2', 'School Admin', false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F3', 'Student', true, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F4', 'Parent',  true, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F5', 'HQ Staff', false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F6', 'Centre Lead', false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F7', 'Teacher Lead', false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F8', 'Centre Manager',  false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F9', 'Centre Staff',  false, now(), now(), '-2147483630'),
  ('01GFMMFRXDZHGVC3YWC7J668F0', 'OpenAPI',  true, now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F1', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F2', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F3', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F4', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F5', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F6', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F7', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F8', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F9', now(), now(), '-2147483630'),
  ('01GFMMFRXD1J2ZWE640JQBMJ0J', '01GFMMFRXDZHGVC3YWC7J668F0', now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GFMMFRXD7KWYFJ6MG22M62M1', 'Teacher', true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M2', 'School Admin', true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M3', 'Student', true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M4', 'Parent',  true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M5', 'HQ Staff', true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M6', 'Centre Lead', true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M7', 'Teacher Lead', true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M8', 'Centre Manager',  true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M9', 'Centre Staff',  true, now(), now(), '-2147483630'),
  ('01GFMMFRXD7KWYFJ6MG22M62M0', 'OpenAPI',  true, now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GFMMFRXDFKR62AVYZVG16S81', '01GFMMFRXD7KWYFJ6MG22M62M1', '01GFMMFRXDZHGVC3YWC7J668F1', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S82', '01GFMMFRXD7KWYFJ6MG22M62M2', '01GFMMFRXDZHGVC3YWC7J668F2', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S83', '01GFMMFRXD7KWYFJ6MG22M62M3', '01GFMMFRXDZHGVC3YWC7J668F3', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S84', '01GFMMFRXD7KWYFJ6MG22M62M4', '01GFMMFRXDZHGVC3YWC7J668F4', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S85', '01GFMMFRXD7KWYFJ6MG22M62M5', '01GFMMFRXDZHGVC3YWC7J668F5', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S86', '01GFMMFRXD7KWYFJ6MG22M62M6', '01GFMMFRXDZHGVC3YWC7J668F6', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S87', '01GFMMFRXD7KWYFJ6MG22M62M7', '01GFMMFRXDZHGVC3YWC7J668F7', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S88', '01GFMMFRXD7KWYFJ6MG22M62M8', '01GFMMFRXDZHGVC3YWC7J668F8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S89', '01GFMMFRXD7KWYFJ6MG22M62M9', '01GFMMFRXDZHGVC3YWC7J668F9', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S80', '01GFMMFRXD7KWYFJ6MG22M62M0', '01GFMMFRXDZHGVC3YWC7J668F0', now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GFMMFRXDFKR62AVYZVG16S81', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S82', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S83', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S84', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S85', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S86', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S87', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S88', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S89', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
  ('01GFMMFRXDFKR62AVYZVG16S80', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630')
  ON CONFLICT DO NOTHING;

------------ Managara High School -------------
--- Location ---
INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01GFMNHQ1WHGRC8AW6K913AM3G','org','Managara High School', '-2147483629', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('01GFMNHQ1WHGRC8AW6K913AM3G', 'Managara High School','01GFMNHQ1WHGRC8AW6K913AM3G',NULL, NULL, NULL, '-2147483629', now(), now(),'01GFMNHQ1WHGRC8AW6K913AM3G') ON CONFLICT DO NOTHING;

--- Organization ---
INSERT INTO organizations (organization_id, tenant_id,               name,       resource_path, domain_name, logo_url, country,      created_at, updated_at, deleted_at)
VALUES                    ('-2147483629',   'withus-managara-hs-2391o', 'Managara High School', '-2147483629', 'managara-hs', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant-managara-hs-logo.png',     'COUNTRY_JP', now(),      now(),      null      ) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483629, 'student-coach-e1e95', 'withus-managara-hs-2391o') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id,   name,        country,      city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES              (-2147483629, 'Managara High School', 'COUNTRY_JP', 1,       1,           null,  false,            now(),      now(),      false,    null,         null,       '-2147483629') ON CONFLICT DO NOTHING;

--- Dynamic form ---
INSERT
	INTO
	public.partner_form_configs (form_config_id,
	partner_id,
	feature_name,
	created_at,
	updated_at,
	deleted_at,
	form_config_data,
	resource_path)
VALUES
		('01GFMNHQ209Z4BTZBGY7X5MPPB',
-2147483629,
'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT',
now(),
now(),
NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_required": true, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_remark", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REMARK"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Homework Submission", "ja": "課題", "vi": "Homework Submission"}, "fallback_language": "ja"}}, "field_id": "homework_submission_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework Status", "ja": "提出状況", "vi": "Homework Status"}, "fallback_language": "ja"}}, "field_id": "homework_submission_status", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "label": {"i18n": {"translations": {"en": "Completed", "ja": "完了", "vi": "Completed"}, "fallback_language": "ja"}}}, {"key": "INCOMPLETE", "label": {"i18n": {"translations": {"en": "Incomplete", "ja": "未完了", "vi": "Incomplete"}, "fallback_language": "ja"}}}], "optionLabelKey": "label"}, "component_config": {"type": "AUTOCOMPLETE"}}], "section_id": "homework_submission_section_id", "section_name": "homework_submission"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Lesson", "ja": "授業", "vi": "Lesson"}, "fallback_language": "ja"}}, "field_id": "lesson_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_view_study_plan_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "LINK_VIEW_STUDY_PLAN"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "追加教材", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "追加課題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "lesson_section_id", "section_name": "lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_section_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "remarks_section_id", "section_name": "remarks"}]}',
'-2147483629')
	ON
CONFLICT ON
CONSTRAINT partner_form_configs_pk DO NOTHING;

--- Role <> User Group ---
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GFMNZZS41E7BEDAYDFHTTTGY', 'master.location.read', now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;
INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01GFMNZZS2HKS7Y1J6EQPGEBC1', 'Teacher', false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC2', 'School Admin', false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC3', 'Student', true, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC4', 'Parent',  true, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC5', 'HQ Staff', false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC6', 'Centre Lead', false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC7', 'Teacher Lead', false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC8', 'Centre Manager',  false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC9', 'Centre Staff',  false, now(), now(), '-2147483629'),
  ('01GFMNZZS2HKS7Y1J6EQPGEBC0', 'OpenAPI',  true, now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC1', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC2', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC3', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC4', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC5', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC6', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC7', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC8', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC9', now(), now(), '-2147483629'),
  ('01GFMNZZS41E7BEDAYDFHTTTGY', '01GFMNZZS2HKS7Y1J6EQPGEBC0', now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M1', 'Teacher', true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M2', 'School Admin', true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M3', 'Student', true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M4', 'Parent',  true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M5', 'HQ Staff', true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M6', 'Centre Lead', true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M7', 'Teacher Lead', true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M8', 'Centre Manager',  true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M9', 'Centre Staff',  true, now(), now(), '-2147483629'),
  ('01GFMNHQ1ZY20VNJCJZQ5EX4M0', 'OpenAPI',  true, now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GFMNZZS44489HGEBQ8W5DT31', '01GFMNHQ1ZY20VNJCJZQ5EX4M1', '01GFMNZZS2HKS7Y1J6EQPGEBC1', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT32', '01GFMNHQ1ZY20VNJCJZQ5EX4M2', '01GFMNZZS2HKS7Y1J6EQPGEBC2', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT33', '01GFMNHQ1ZY20VNJCJZQ5EX4M3', '01GFMNZZS2HKS7Y1J6EQPGEBC3', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT34', '01GFMNHQ1ZY20VNJCJZQ5EX4M4', '01GFMNZZS2HKS7Y1J6EQPGEBC4', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT35', '01GFMNHQ1ZY20VNJCJZQ5EX4M5', '01GFMNZZS2HKS7Y1J6EQPGEBC5', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT36', '01GFMNHQ1ZY20VNJCJZQ5EX4M6', '01GFMNZZS2HKS7Y1J6EQPGEBC6', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT37', '01GFMNHQ1ZY20VNJCJZQ5EX4M7', '01GFMNZZS2HKS7Y1J6EQPGEBC7', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT38', '01GFMNHQ1ZY20VNJCJZQ5EX4M8', '01GFMNZZS2HKS7Y1J6EQPGEBC8', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT39', '01GFMNHQ1ZY20VNJCJZQ5EX4M9', '01GFMNZZS2HKS7Y1J6EQPGEBC9', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT30', '01GFMNHQ1ZY20VNJCJZQ5EX4M0', '01GFMNZZS2HKS7Y1J6EQPGEBC0', now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GFMNZZS44489HGEBQ8W5DT31', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT32', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT33', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT34', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT35', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT36', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT37', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT38', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT39', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
  ('01GFMNZZS44489HGEBQ8W5DT30', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629')
  ON CONFLICT DO NOTHING;
