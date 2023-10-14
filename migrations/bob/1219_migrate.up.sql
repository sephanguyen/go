INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q8M2','org','E2E HCM', '-2147483638', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q8N2', 'E2E HCM','01FR4M51XJY9E77GSN4QZ1Q8M2',NULL, NULL, NULL, '-2147483638', now(), now(),'01FR4M51XJY9E77GSN4QZ1Q8N2') ON CONFLICT DO NOTHING;

INSERT INTO public.schools
(school_id, "name", country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at)
VALUES(-2147483638, 'E2E HCM', 'COUNTRY_JP', 1, 1, NULL, false, now(), now(), false, NULL, NULL) ON CONFLICT DO NOTHING;

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
		('01FTCP1VPV85CV5C5RH7FKQ2BW',
-2147483638,
'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT',
now(),
now(),
NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_required": true, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_remark", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REMARK"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Homework Submission", "ja": "課題", "vi": "Homework Submission"}, "fallback_language": "ja"}}, "field_id": "homework_submission_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework Status", "ja": "提出状況", "vi": "Homework Status"}, "fallback_language": "ja"}}, "field_id": "homework_submission_status", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "label": {"i18n": {"translations": {"en": "Completed", "ja": "完了", "vi": "Completed"}, "fallback_language": "ja"}}}, {"key": "INCOMPLETE", "label": {"i18n": {"translations": {"en": "Incomplete", "ja": "未完了", "vi": "Incomplete"}, "fallback_language": "ja"}}}], "optionLabelKey": "label"}, "component_config": {"type": "AUTOCOMPLETE"}}], "section_id": "homework_submission_section_id", "section_name": "homework_submission"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Lesson", "ja": "授業", "vi": "Lesson"}, "fallback_language": "ja"}}, "field_id": "lesson_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_view_study_plan_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "LINK_VIEW_STUDY_PLAN"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "追加教材", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "追加課題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "lesson_section_id", "section_name": "lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_section_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "remarks_section_id", "section_name": "remarks"}]}',
'-2147483638')
	ON
CONFLICT ON
CONSTRAINT partner_form_configs_pk DO NOTHING;



INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q8M3','org','Manabie Demo', '-2147483637', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q8N3', 'Manabie Demo','01FR4M51XJY9E77GSN4QZ1Q8M3',NULL, NULL, NULL, '-2147483637', now(), now(),'01FR4M51XJY9E77GSN4QZ1Q8N3') ON CONFLICT DO NOTHING;


INSERT INTO public.schools
(school_id, "name", country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at)
VALUES(-2147483637, 'Manabie Demo', 'COUNTRY_JP', 1, 1, NULL, false, now(), now(), false, NULL, NULL) ON CONFLICT DO NOTHING;

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
		('01FTC31VPV85CV5C5RH7FKQ2BW',
-2147483637,
'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT',
now(),
now(),
NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_required": true, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_remark", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REMARK"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Homework Submission", "ja": "課題", "vi": "Homework Submission"}, "fallback_language": "ja"}}, "field_id": "homework_submission_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework Status", "ja": "提出状況", "vi": "Homework Status"}, "fallback_language": "ja"}}, "field_id": "homework_submission_status", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "label": {"i18n": {"translations": {"en": "Completed", "ja": "完了", "vi": "Completed"}, "fallback_language": "ja"}}}, {"key": "INCOMPLETE", "label": {"i18n": {"translations": {"en": "Incomplete", "ja": "未完了", "vi": "Incomplete"}, "fallback_language": "ja"}}}], "optionLabelKey": "label"}, "component_config": {"type": "AUTOCOMPLETE"}}], "section_id": "homework_submission_section_id", "section_name": "homework_submission"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Lesson", "ja": "授業", "vi": "Lesson"}, "fallback_language": "ja"}}, "field_id": "lesson_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_view_study_plan_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "LINK_VIEW_STUDY_PLAN"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "追加教材", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "追加課題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "lesson_section_id", "section_name": "lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_section_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "remarks_section_id", "section_name": "remarks"}]}',
'-2147483637')
	ON
CONFLICT ON
CONSTRAINT partner_form_configs_pk DO NOTHING;
