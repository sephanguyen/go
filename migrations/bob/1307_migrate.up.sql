--- Org and Location ---
INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01GDWSMJS6APH4SX2NP5NFWHG5','org','Eishinkan', '-2147483631', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('01GDWSMJS6APH4SX2NP5NFWHG5', 'Eishinkan','01GDWSMJS6APH4SX2NP5NFWHG5',NULL, NULL, NULL, '-2147483631', now(), now(),'01GDWSMJS6APH4SX2NP5NFWHG5') ON CONFLICT DO NOTHING;

INSERT INTO organizations (organization_id, tenant_id,               name,       resource_path, domain_name, logo_url, country,      created_at, updated_at, deleted_at)
VALUES                    ('-2147483631',   'eishinkan-de5sd', 'Eishinkan', '-2147483631', 'eishinkan-group', '',     'COUNTRY_JP', now(),      now(),      null      ) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(-2147483631, 'student-coach-e1e95', 'eishinkan-de5sd') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id,   name,        country,      city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES              (-2147483631, 'Eishinkan', 'COUNTRY_JP', 1,       1,           null,  false,            now(),      now(),      false,    null,         null,       '-2147483631') ON CONFLICT DO NOTHING;

INSERT INTO
	public.partner_form_configs (form_config_id,
	partner_id,
	feature_name,
	created_at,
	updated_at,
	deleted_at,
	form_config_data,
	resource_path)
VALUES
		('81FTCP1VPV85CV5C5RH7FKQ2B5',
-2147483631,
'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT',
now(),
now(),
NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_required": true, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_remark", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REMARK"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Homework Submission", "ja": "課題", "vi": "Homework Submission"}, "fallback_language": "ja"}}, "field_id": "homework_submission_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework Status", "ja": "提出状況", "vi": "Homework Status"}, "fallback_language": "ja"}}, "field_id": "homework_submission_status", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "label": {"i18n": {"translations": {"en": "Completed", "ja": "完了", "vi": "Completed"}, "fallback_language": "ja"}}}, {"key": "INCOMPLETE", "label": {"i18n": {"translations": {"en": "Incomplete", "ja": "未完了", "vi": "Incomplete"}, "fallback_language": "ja"}}}], "optionLabelKey": "label"}, "component_config": {"type": "AUTOCOMPLETE"}}], "section_id": "homework_submission_section_id", "section_name": "homework_submission"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Lesson", "ja": "授業", "vi": "Lesson"}, "fallback_language": "ja"}}, "field_id": "lesson_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_view_study_plan_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "LINK_VIEW_STUDY_PLAN"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "追加教材", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "追加課題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "lesson_section_id", "section_name": "lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_section_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "remarks_section_id", "section_name": "remarks"}]}',
'-2147483631')
	ON
CONFLICT ON
CONSTRAINT partner_form_configs_pk DO NOTHING;

--- Permission and Role ---
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', 'master.location.read', now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;
INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01GDWSMJS45TK897ZA6TN2N671', 'Teacher', false, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N672', 'School Admin', false, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N673', 'Student', true, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N674', 'Parent',  true, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N675', 'HQ Staff', false, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N676', 'Centre Lead', false, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N677', 'Teacher Lead', false, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N678', 'Centre Manager',  false, now(), now(), '-2147483631'),
  ('01GDWSMJS45TK897ZA6TN2N679', 'Centre Staff',  false, now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N671', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N672', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N673', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N674', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N675', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N676', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N677', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N678', now(), now(), '-2147483631'),
  ('01G8T4EYFQTSHV61G8Q2T3XPVM', '01GDWSMJS45TK897ZA6TN2N679', now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GDWSMJS5NPM7G49841NFTCC31', 'Teacher', true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC32', 'School Admin', true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC33', 'Student', true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC34', 'Parent',  true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC35', 'HQ Staff', true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC36', 'Centre Lead', true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC37', 'Teacher Lead', true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC38', 'Centre Manager',  true, now(), now(), '-2147483631'),
  ('01GDWSMJS5NPM7G49841NFTCC39', 'Centre Staff',  true, now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GDWSMJS5TTFGSZJTR49VPPV01', '01GDWSMJS5NPM7G49841NFTCC31', '01GDWSMJS45TK897ZA6TN2N671', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV02', '01GDWSMJS5NPM7G49841NFTCC32', '01GDWSMJS45TK897ZA6TN2N672', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV03', '01GDWSMJS5NPM7G49841NFTCC33', '01GDWSMJS45TK897ZA6TN2N673', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV04', '01GDWSMJS5NPM7G49841NFTCC34', '01GDWSMJS45TK897ZA6TN2N674', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV05', '01GDWSMJS5NPM7G49841NFTCC35', '01GDWSMJS45TK897ZA6TN2N675', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV06', '01GDWSMJS5NPM7G49841NFTCC36', '01GDWSMJS45TK897ZA6TN2N676', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV07', '01GDWSMJS5NPM7G49841NFTCC37', '01GDWSMJS45TK897ZA6TN2N677', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV08', '01GDWSMJS5NPM7G49841NFTCC38', '01GDWSMJS45TK897ZA6TN2N678', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV09', '01GDWSMJS5NPM7G49841NFTCC39', '01GDWSMJS45TK897ZA6TN2N679', now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GDWSMJS5TTFGSZJTR49VPPV01', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV02', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV03', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV04', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV05', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV06', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV07', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV08', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
  ('01GDWSMJS5TTFGSZJTR49VPPV09', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631')
  ON CONFLICT DO NOTHING;