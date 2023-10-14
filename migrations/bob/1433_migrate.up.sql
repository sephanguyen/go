------------ E2E Architecture -------------
--- Location ---
INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('911FLMNMYA6SKTTT44HWE2E100','org','E2E Architecture', '100000', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('911FLMNMYA6SKTTT44HWE2E100', 'E2E Architecture','911FLMNMYA6SKTTT44HWE2E100',NULL, NULL, NULL, '100000', now(), now(),'911FLMNMYA6SKTTT44HWE2E100') ON CONFLICT DO NOTHING;

--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES ('100000', 'e2e-architecture-29vl6', 'E2E Architecture', '100000', 'e2e-architecture', '', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

INSERT INTO public.organization_auths
(organization_id, auth_project_id, auth_tenant_id)
VALUES(100000, 'student-coach-e1e95', 'e2e-architecture-29vl6') ON CONFLICT DO NOTHING;

INSERT INTO schools ( school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
VALUES (100000, 'E2E Architecture', 'COUNTRY_JP', 1, 1, null, false, now(), now(), false, null, null, '100000') ON CONFLICT DO NOTHING;

--- Dynamic form ---
INSERT INTO public.partner_form_configs (form_config_id,
    partner_id,
    feature_name,
    created_at,
    updated_at,
    deleted_at,
    form_config_data,
    resource_path)
VALUES ('911FLMNMYAZHGVC3YWCFORM100', 100000, 'FORM_CONFIG_LESSON_REPORT_IND_UPDATE', now(), now(), NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_notice", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_NOTICE"}}, {"field_id": "attendance_reason", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REASON"}}, {"field_id": "attendance_note", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_NOTE"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "This Lesson", "ja": "今回の授業", "vi": "This Lesson"}, "fallback_language": "ja"}}, "field_id": "this_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 9, "xs": 9}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "授業内容", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework Completion", "ja": "宿題提出", "vi": "Homework Completion"}, "fallback_language": "ja"}}, "field_id": "homework_completion", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "icon": "CircleOutlined"}, {"key": "IN_PROGRESS", "icon": "ChangeHistoryOutlined"}, {"key": "INCOMPLETE", "icon": "CloseOutlined"}], "valueKey": "key", "optionIconLabelKey": "icon"}, "component_config": {"type": "SELECT_ICON"}}, {"label": {"i18n": {"translations": {"en": "In-Lesson Quiz", "ja": "授業内クイズ", "vi": "In-Lesson Quiz"}, "fallback_language": "ja"}}, "field_id": "in_lesson_quiz", "value_type": "VALUE_TYPE_INT", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "TEXT_FIELD_PERCENTAGE"}}, {"label": {"i18n": {"translations": {"en": "Understanding", "ja": "理解度", "vi": "Understanding"}, "fallback_language": "ja"}}, "field_id": "lesson_understanding", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "A", "label": {"i18n": {"translations": {"en": "A", "ja": "A", "vi": "A"}, "fallback_language": "ja"}}}, {"key": "B", "label": {"i18n": {"translations": {"en": "B", "ja": "B", "vi": "B"}, "fallback_language": "ja"}}}, {"key": "C", "label": {"i18n": {"translations": {"en": "C", "ja": "C", "vi": "C"}, "fallback_language": "ja"}}}, {"key": "D", "label": {"i18n": {"translations": {"en": "D", "ja": "D", "vi": "D"}, "fallback_language": "ja"}}}], "disableCloseOnSelect": false}, "component_config": {"type": "AUTOCOMPLETE_MANA_UI"}}], "section_id": "this_lesson_section_id", "section_name": "this_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Next Lesson", "ja": "次回の授業", "vi": "Next Lesson"}, "fallback_language": "ja"}}, "field_id": "next_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "宿題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Announcement", "ja": "お知らせ", "vi": "Announcement"}, "fallback_language": "ja"}}, "field_id": "lesson_announcement", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "next_lesson_section_id", "section_name": "next_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "特筆事項", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "remark_internal", "value_type": "VALUE_TYPE_STRING", "is_internal": true, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA", "question_mark": {"message": {"i18n": {"translations": {"en": "This is an internal memo, it will not be shared with parents", "ja": "これは社内用メモです。保護者には共有されません", "vi": "This is an internal memo, it will not be shared with parents"}, "fallback_language": "ja"}}}}}], "section_id": "remark_section_id", "section_name": "remark"}]}',
'100000') on CONFLICT on CONSTRAINT partner_form_configs_pk DO NOTHING;

INSERT INTO public.partner_form_configs (form_config_id,
    partner_id,
    feature_name,
    created_at,
    updated_at,
    deleted_at,
    form_config_data,
    resource_path)
VALUES ('911FLMNMYAZHGVCFORMGROUP100', 100000, 'FEATURE_NAME_GROUP_LESSON_REPORT', now(), now(), NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "This Lesson", "ja": "今回の授業", "vi": "This Lesson"}, "fallback_language": "ja"}}, "field_id": "this_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 10, "xs": 10}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 2, "xs": 2}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "授業内容", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "content", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "lesson_remark", "value_type": "VALUE_TYPE_STRING", "is_internal": true, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA", "question_mark": {"message": {"i18n": {"translations": {"en": "This is an internal memo, it will not be shared with parents", "ja": "これは社内用メモです。保護者には共有されません", "vi": "This is an internal memo, it will not be shared with parents"}, "fallback_language": "ja"}}}}}], "section_id": "this_lesson_id", "section_name": "this_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Next Lesson", "ja": "次回の授業", "vi": "Next Lesson"}, "fallback_language": "ja"}}, "field_id": "next_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "宿題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "homework", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Announcement", "ja": "お知らせ", "vi": "Announcement"}, "fallback_language": "ja"}}, "field_id": "announcement", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "next_lesson_id", "section_name": "next_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Student List", "ja": "生徒リスト", "vi": "Student List"}, "fallback_language": "ja"}}, "field_id": "student_list_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TOGGLE_TABLE_TITLE"}}, {"field_id": "student_list_tables", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"dynamicFields": [], "toggleButtons": [{"label": {"i18n": {"translations": {"en": "Performance", "ja": "成績", "vi": "Performance"}, "fallback_language": "ja"}}, "field_id": "performance"}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "備考", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "remark"}]}, "component_config": {"type": "TOGGLE_TABLE"}}, {"label": {"i18n": {"translations": {"en": "Homework Completion", "ja": "宿題提出", "vi": "Homework Completion"}, "fallback_language": "ja"}}, "field_id": "homework_completion", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"table_size": {"width": "22%"}}, "component_props": {"options": [{"key": "COMPLETED", "icon": "CircleOutlined"}, {"key": "IN_PROGRESS", "icon": "ChangeHistoryOutlined"}, {"key": "INCOMPLETE", "icon": "CloseOutlined"}], "variant": "body2", "valueKey": "key", "placeholder": {"i18n": {"translations": {"en": "Homework Completion", "ja": "宿題提出", "vi": "Homework Completion"}, "fallback_language": "ja"}}, "optionIconLabelKey": "icon"}, "component_config": {"type": "SELECT_ICON", "table_key": "performance", "has_bulk_action": true}}, {"label": {"i18n": {"translations": {"en": "In-lesson Quiz", "ja": "授業内クイズ", "vi": "In-lesson Quiz"}, "fallback_language": "ja"}}, "field_id": "in_lesson_quiz", "value_type": "VALUE_TYPE_INT", "is_internal": false, "is_required": false, "display_config": {"table_size": {"width": "22%"}}, "component_props": {"variant": "body2", "placeholder": {"i18n": {"translations": {"en": "In-lesson Quiz", "ja": "授業内クイズ", "vi": "In-lesson Quiz"}, "fallback_language": "ja"}}}, "component_config": {"type": "TEXT_FIELD_PERCENTAGE", "table_key": "performance", "has_bulk_action": true}}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "student_remark", "value_type": "VALUE_TYPE_STRING", "is_internal": true, "is_required": false, "display_config": {"table_size": {"width": "70%"}}, "component_props": {"variant": "body2", "placeholder": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}}, "component_config": {"type": "TEXT_FIELD", "table_key": "remark", "question_mark": {"message": {"i18n": {"translations": {"en": "This is an internal memo, it will not be shared with parents", "ja": "これは社内用メモです。保護者には共有されません", "vi": "This is an internal memo, it will not be shared with parents"}, "fallback_language": "ja"}}}, "has_bulk_action": false}}], "section_id": "student_list_id", "section_name": "student_list"}]}',
'100000') on CONFLICT on CONSTRAINT partner_form_configs_pk DO NOTHING;

--- Role <> User Group ---
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('911FLMNMYA1J2ZWE640JQBMJ0J', 'master.location.read', now(), now(), '100000')
  ON CONFLICT DO NOTHING;
INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('911FLMNMYAZHGVC3YWC7J668F1', 'Teacher', false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F2', 'School Admin', false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F3', 'Student', true, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F4', 'Parent',  true, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F5', 'HQ Staff', false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F6', 'Centre Lead', false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F7', 'Teacher Lead', false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F8', 'Centre Manager',  false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F9', 'Centre Staff',  false, now(), now(), '100000'),
  ('911FLMNMYAZHGVC3YWC7J668F0', 'OpenAPI',  true, now(), now(), '100000')
  ON CONFLICT DO NOTHING;
INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F1', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F2', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F3', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F4', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F5', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F6', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F7', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F8', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F9', now(), now(), '100000'),
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '911FLMNMYAZHGVC3YWC7J668F0', now(), now(), '100000')
  ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('911FLMNMYA7KWYFJ6MG22M62M1', 'Teacher', true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M2', 'School Admin', true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M3', 'Student', true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M4', 'Parent',  true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M5', 'HQ Staff', true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M6', 'Centre Lead', true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M7', 'Teacher Lead', true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M8', 'Centre Manager',  true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M9', 'Centre Staff',  true, now(), now(), '100000'),
  ('911FLMNMYA7KWYFJ6MG22M62M0', 'OpenAPI',  true, now(), now(), '100000')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('911FLMNMYAFKR62AVYZVG16S81', '911FLMNMYA7KWYFJ6MG22M62M1', '911FLMNMYAZHGVC3YWC7J668F1', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S82', '911FLMNMYA7KWYFJ6MG22M62M2', '911FLMNMYAZHGVC3YWC7J668F2', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S83', '911FLMNMYA7KWYFJ6MG22M62M3', '911FLMNMYAZHGVC3YWC7J668F3', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S84', '911FLMNMYA7KWYFJ6MG22M62M4', '911FLMNMYAZHGVC3YWC7J668F4', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S85', '911FLMNMYA7KWYFJ6MG22M62M5', '911FLMNMYAZHGVC3YWC7J668F5', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S86', '911FLMNMYA7KWYFJ6MG22M62M6', '911FLMNMYAZHGVC3YWC7J668F6', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S87', '911FLMNMYA7KWYFJ6MG22M62M7', '911FLMNMYAZHGVC3YWC7J668F7', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S88', '911FLMNMYA7KWYFJ6MG22M62M8', '911FLMNMYAZHGVC3YWC7J668F8', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S89', '911FLMNMYA7KWYFJ6MG22M62M9', '911FLMNMYAZHGVC3YWC7J668F9', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S80', '911FLMNMYA7KWYFJ6MG22M62M0', '911FLMNMYAZHGVC3YWC7J668F0', now(), now(), '100000')
  ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('911FLMNMYAFKR62AVYZVG16S81', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S82', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S83', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S84', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S85', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S86', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S87', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S88', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S89', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000'),
  ('911FLMNMYAFKR62AVYZVG16S80', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000')
  ON CONFLICT DO NOTHING;
