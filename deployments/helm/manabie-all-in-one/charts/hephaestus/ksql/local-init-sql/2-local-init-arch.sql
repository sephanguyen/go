
\connect bob;

--- Dynamic form ---
INSERT INTO public.partner_form_configs (form_config_id,
    partner_id,
    feature_name,
    created_at,
    updated_at,
    deleted_at,
    form_config_data,
    resource_path)
VALUES ('911FLMNMYAZHGVC3YWCFORM2147483644', -2147483644, 'FORM_CONFIG_LESSON_REPORT_IND_UPDATE', now(), now(), NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_notice", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_NOTICE"}}, {"field_id": "attendance_reason", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REASON"}}, {"field_id": "attendance_note", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_NOTE"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "This Lesson", "ja": "今回の授業", "vi": "This Lesson"}, "fallback_language": "ja"}}, "field_id": "this_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 9, "xs": 9}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "授業内容", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework Completion", "ja": "宿題提出", "vi": "Homework Completion"}, "fallback_language": "ja"}}, "field_id": "homework_completion", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "icon": "CircleOutlined"}, {"key": "IN_PROGRESS", "icon": "ChangeHistoryOutlined"}, {"key": "INCOMPLETE", "icon": "CloseOutlined"}], "valueKey": "key", "optionIconLabelKey": "icon"}, "component_config": {"type": "SELECT_ICON"}}, {"label": {"i18n": {"translations": {"en": "In-Lesson Quiz", "ja": "授業内クイズ", "vi": "In-Lesson Quiz"}, "fallback_language": "ja"}}, "field_id": "in_lesson_quiz", "value_type": "VALUE_TYPE_INT", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_config": {"type": "TEXT_FIELD_PERCENTAGE"}}, {"label": {"i18n": {"translations": {"en": "Understanding", "ja": "理解度", "vi": "Understanding"}, "fallback_language": "ja"}}, "field_id": "lesson_understanding", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "A", "label": {"i18n": {"translations": {"en": "A", "ja": "A", "vi": "A"}, "fallback_language": "ja"}}}, {"key": "B", "label": {"i18n": {"translations": {"en": "B", "ja": "B", "vi": "B"}, "fallback_language": "ja"}}}, {"key": "C", "label": {"i18n": {"translations": {"en": "C", "ja": "C", "vi": "C"}, "fallback_language": "ja"}}}, {"key": "D", "label": {"i18n": {"translations": {"en": "D", "ja": "D", "vi": "D"}, "fallback_language": "ja"}}}], "disableCloseOnSelect": false}, "component_config": {"type": "AUTOCOMPLETE_MANA_UI"}}], "section_id": "this_lesson_section_id", "section_name": "this_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Next Lesson", "ja": "次回の授業", "vi": "Next Lesson"}, "fallback_language": "ja"}}, "field_id": "next_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "宿題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Announcement", "ja": "お知らせ", "vi": "Announcement"}, "fallback_language": "ja"}}, "field_id": "lesson_announcement", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "next_lesson_section_id", "section_name": "next_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "特筆事項", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "remark_internal", "value_type": "VALUE_TYPE_STRING", "is_internal": true, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA", "question_mark": {"message": {"i18n": {"translations": {"en": "This is an internal memo, it will not be shared with parents", "ja": "これは社内用メモです。保護者には共有されません", "vi": "This is an internal memo, it will not be shared with parents"}, "fallback_language": "ja"}}}}}], "section_id": "remark_section_id", "section_name": "remark"}]}',
'-2147483644') on CONFLICT on CONSTRAINT partner_form_configs_pk DO NOTHING;

INSERT INTO public.partner_form_configs (form_config_id,
    partner_id,
    feature_name,
    created_at,
    updated_at,
    deleted_at,
    form_config_data,
    resource_path)
VALUES ('911FLMNMYAZHGVCFORMGROUP2147483644', -2147483644, 'FEATURE_NAME_GROUP_LESSON_REPORT', now(), now(), NULL,
'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "This Lesson", "ja": "今回の授業", "vi": "This Lesson"}, "fallback_language": "ja"}}, "field_id": "this_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 10, "xs": 10}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 2, "xs": 2}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "授業内容", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "content", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "lesson_remark", "value_type": "VALUE_TYPE_STRING", "is_internal": true, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA", "question_mark": {"message": {"i18n": {"translations": {"en": "This is an internal memo, it will not be shared with parents", "ja": "これは社内用メモです。保護者には共有されません", "vi": "This is an internal memo, it will not be shared with parents"}, "fallback_language": "ja"}}}}}], "section_id": "this_lesson_id", "section_name": "this_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Next Lesson", "ja": "次回の授業", "vi": "Next Lesson"}, "fallback_language": "ja"}}, "field_id": "next_lesson_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "宿題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "homework", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": true, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Announcement", "ja": "お知らせ", "vi": "Announcement"}, "fallback_language": "ja"}}, "field_id": "announcement", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "next_lesson_id", "section_name": "next_lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Student List", "ja": "生徒リスト", "vi": "Student List"}, "fallback_language": "ja"}}, "field_id": "student_list_label", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"variant": "subtitle1"}, "component_config": {"type": "TOGGLE_TABLE_TITLE"}}, {"field_id": "student_list_tables", "value_type": "VALUE_TYPE_NULL", "is_internal": false, "is_required": false, "display_config": {"grid_size": {"md": 12, "xs": 12}}, "component_props": {"dynamicFields": [], "toggleButtons": [{"label": {"i18n": {"translations": {"en": "Performance", "ja": "成績", "vi": "Performance"}, "fallback_language": "ja"}}, "field_id": "performance"}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "備考", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "remark"}]}, "component_config": {"type": "TOGGLE_TABLE"}}, {"label": {"i18n": {"translations": {"en": "Homework Completion", "ja": "宿題提出", "vi": "Homework Completion"}, "fallback_language": "ja"}}, "field_id": "homework_completion", "value_type": "VALUE_TYPE_STRING", "is_internal": false, "is_required": false, "display_config": {"table_size": {"width": "22%"}}, "component_props": {"options": [{"key": "COMPLETED", "icon": "CircleOutlined"}, {"key": "IN_PROGRESS", "icon": "ChangeHistoryOutlined"}, {"key": "INCOMPLETE", "icon": "CloseOutlined"}], "variant": "body2", "valueKey": "key", "placeholder": {"i18n": {"translations": {"en": "Homework Completion", "ja": "宿題提出", "vi": "Homework Completion"}, "fallback_language": "ja"}}, "optionIconLabelKey": "icon"}, "component_config": {"type": "SELECT_ICON", "table_key": "performance", "has_bulk_action": true}}, {"label": {"i18n": {"translations": {"en": "In-lesson Quiz", "ja": "授業内クイズ", "vi": "In-lesson Quiz"}, "fallback_language": "ja"}}, "field_id": "in_lesson_quiz", "value_type": "VALUE_TYPE_INT", "is_internal": false, "is_required": false, "display_config": {"table_size": {"width": "22%"}}, "component_props": {"variant": "body2", "placeholder": {"i18n": {"translations": {"en": "In-lesson Quiz", "ja": "授業内クイズ", "vi": "In-lesson Quiz"}, "fallback_language": "ja"}}}, "component_config": {"type": "TEXT_FIELD_PERCENTAGE", "table_key": "performance", "has_bulk_action": true}}, {"label": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}, "field_id": "student_remark", "value_type": "VALUE_TYPE_STRING", "is_internal": true, "is_required": false, "display_config": {"table_size": {"width": "70%"}}, "component_props": {"variant": "body2", "placeholder": {"i18n": {"translations": {"en": "Remark", "ja": "特筆事項", "vi": "Remark"}, "fallback_language": "ja"}}}, "component_config": {"type": "TEXT_FIELD", "table_key": "remark", "question_mark": {"message": {"i18n": {"translations": {"en": "This is an internal memo, it will not be shared with parents", "ja": "これは社内用メモです。保護者には共有されません", "vi": "This is an internal memo, it will not be shared with parents"}, "fallback_language": "ja"}}}, "has_bulk_action": false}}], "section_id": "student_list_id", "section_name": "student_list"}]}',
'-2147483644') on CONFLICT on CONSTRAINT partner_form_configs_pk DO NOTHING;


INSERT INTO public.location_types
(location_type_id, "name", display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, "level")
VALUES('01GV2MB7J1XH9JN4RGA2Y06JZC', 'brand', 'Brand', 'org', '01FR4M51XJY9E77GSN4QZ1Q9M5', '2023-03-09 14:28:39.679', '2023-03-09 14:28:39.489', NULL, '-2147483644', false, 1) ON CONFLICT DO NOTHING;
INSERT INTO public.location_types
(location_type_id, "name", display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, "level")
VALUES('01GV2MB7K7P24Q80HWSXNB0RN8', 'center', 'Center', 'brand', '01GV2MB7J1XH9JN4RGA2Y06JZC', '2023-03-09 14:28:39.679', '2023-03-09 14:28:39.527', NULL, '-2147483644', false, 2) ON CONFLICT DO NOTHING;


INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBJR5PGRHC601V0NMHKGY', 'Brand 1', '2023-03-09 14:28:50.933', '2023-03-09 14:28:51.563', NULL, '-2147483644', '01GV2MB7J1XH9JN4RGA2Y06JZC', '1', NULL, '01FR4M51XJY9E77GSN4QZ1Q9N5', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJR5PGRHC601V0NMHKGY') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBJRW9AS88X5S4C5DXCE7', 'Center 2', '2023-03-09 14:28:50.933', '2023-03-09 14:28:52.563', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '2', '1', '01GV2MBJR5PGRHC601V0NMHKGY', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJR5PGRHC601V0NMHKGY/01GV2MBJRW9AS88X5S4C5DXCE7') ON CONFLICT DO NOTHING;


INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBJV336BV73298ED0ND2X', 'Center 3', '2023-03-09 14:28:50.933', '2023-03-09 14:28:53.563', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '3', '1', '01GV2MBJR5PGRHC601V0NMHKGY', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJR5PGRHC601V0NMHKGY/01GV2MBJV336BV73298ED0ND2X') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBJX12HR7N1CTF76CR7VH', 'Center 4', '2023-03-09 14:28:50.933', '2023-03-09 14:28:54.563', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '4', '1', '01GV2MBJR5PGRHC601V0NMHKGY', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJR5PGRHC601V0NMHKGY/01GV2MBJX12HR7N1CTF76CR7VH') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBJZW0CA66KXAHJTZJ86X', 'Brand 1', '2023-03-09 14:28:50.933', '2023-03-09 14:28:55.563', NULL, '-2147483644', '01GV2MB7J1XH9JN4RGA2Y06JZC', '5', NULL, '01FR4M51XJY9E77GSN4QZ1Q9N5', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJZW0CA66KXAHJTZJ86X') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBK19NHHG2PV1SB4JHTEH', 'Center 5', '2023-03-09 14:28:50.933', '2023-03-09 14:28:56.563', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '6', '5', '01GV2MBJZW0CA66KXAHJTZJ86X', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJZW0CA66KXAHJTZJ86X/01GV2MBK19NHHG2PV1SB4JHTEH') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBK3ZCQEJ711H3EW0VEQY', 'Center 6', '2023-03-09 14:28:50.933', '2023-03-09 14:28:57.564', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '7', '5', '01GV2MBJZW0CA66KXAHJTZJ86X', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJZW0CA66KXAHJTZJ86X/01GV2MBK3ZCQEJ711H3EW0VEQY') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBK4PP8M0V7JB4WGR315G', 'Center 7', '2023-03-09 14:28:50.933', '2023-03-09 14:28:58.564', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '8', '5', '01GV2MBJZW0CA66KXAHJTZJ86X', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBJZW0CA66KXAHJTZJ86X/01GV2MBK4PP8M0V7JB4WGR315G') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBK66A6SY7V0TVA8KWJHG', 'Brand 2', '2023-03-09 14:28:50.933', '2023-03-09 14:28:59.564', NULL, '-2147483644', '01GV2MB7J1XH9JN4RGA2Y06JZC', '9', NULL, '01FR4M51XJY9E77GSN4QZ1Q9N5', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBK66A6SY7V0TVA8KWJHG') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBK805A0QREYJSKS92R6A', 'Center 8', '2023-03-09 14:28:50.933', '2023-03-09 14:29:00.564', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '10', '9', '01GV2MBK66A6SY7V0TVA8KWJHG', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBK66A6SY7V0TVA8KWJHG/01GV2MBK805A0QREYJSKS92R6A') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBK94A3Q1184KK8KQJCB7', 'Center 9', '2023-03-09 14:28:50.933', '2023-03-09 14:29:01.564', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '11', '9', '01GV2MBK66A6SY7V0TVA8KWJHG', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBK66A6SY7V0TVA8KWJHG/01GV2MBK94A3Q1184KK8KQJCB7') ON CONFLICT DO NOTHING;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV2MBKA50D3ZY095W69RCX0S', 'Center 10', '2023-03-09 14:28:50.933', '2023-03-09 14:29:02.564', NULL, '-2147483644', '01GV2MB7K7P24Q80HWSXNB0RN8', '12', '9', '01GV2MBK66A6SY7V0TVA8KWJHG', false, '01FR4M51XJY9E77GSN4QZ1Q9N5/01GV2MBK66A6SY7V0TVA8KWJHG/01GV2MBKA50D3ZY095W69RCX0S') ON CONFLICT DO NOTHING;

-- init locations & location types for Manabie
-- 01FR4M51XJY9E77GSN4QZ1Q9M1 is Manabie location type
INSERT INTO public.location_types
  (location_type_id, name, display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, level)
VALUES
  ('location-type-id-1', 'branch', 'branch', 'Manabie Org', '01FR4M51XJY9E77GSN4QZ1Q9M1', NOW(), NOW(), NULL, '-2147483648', false, 1),
  ('location-type-id-2', 'center', 'center', 'branch',      'location-type-id-1',         NOW(), NOW(), NULL, '-2147483648', false, 2)
  ON CONFLICT DO NOTHING;

INSERT INTO public.location_types
(location_type_id, name, display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, level)
VALUES
  ('-2147483635_location-type-id-1', 'branch', 'branch', 'KEC Demo Org', '01FR4M51XJY9E77GSN4QZ1Q8M4', NOW(), NOW(), NULL, '-2147483635', false, 1),
  ('-2147483635_location-type-id-2', 'center', 'center', 'branch', '-2147483635_location-type-id-1', NOW(), NOW(), NULL, '-2147483635', false, 2)
  ON CONFLICT DO NOTHING;

-- 01FR4M51XJY9E77GSN4QZ1Q8N4 is KEC Demo location
INSERT INTO public.locations
  (location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES
  ('-2147483648_location-id-1', '[Manabie] location-id-1',  NOW(),  NOW(), NULL, '-2147483648', 'location-type-id-1', 'location-id-1', NULL, '01FR4M51XJY9E77GSN4QZ1Q9N1', false, '01FR4M51XJY9E77GSN4QZ1Q9N1/-2147483648_location-id-1'),
  ('-2147483648_location-id-2', '[Manabie] location-id-2',  NOW(),  NOW(), NULL, '-2147483648', 'location-type-id-2', 'location-id-2', '-2147483648_location-id-1', '-2147483648_location-id-1', false, '01FR4M51XJY9E77GSN4QZ1Q9N1/-2147483648_location-id-1/-2147483648_location-id-2'),
  ('-2147483648_location-id-3', '[Manabie] location-id-3',  NOW(),  NOW(), NULL, '-2147483648', 'location-type-id-2', 'location-id-3', '-2147483648_location-id-1', '-2147483648_location-id-1', false, '01FR4M51XJY9E77GSN4QZ1Q9N1/-2147483648_location-id-1/-2147483648_location-id-3'),
  ('-2147483648_location-id-4', '[Manabie] location-id-4',  NOW(),  NOW(), NULL, '-2147483648', 'location-type-id-2', 'location-id-4', '-2147483648_location-id-1', '-2147483648_location-id-1', false, '01FR4M51XJY9E77GSN4QZ1Q9N1/-2147483648_location-id-1/-2147483648_location-id-4'),
  ('-2147483635_location-id-1', '[KEC-Demo] location-id-1',  NOW(),  NOW(), NULL, '-2147483635', '-2147483635_location-type-id-1', 'location-id-1', NULL, '01FR4M51XJY9E77GSN4QZ1Q8N4', false, '01FR4M51XJY9E77GSN4QZ1Q8N4/-2147483635_location-id-1'),
  ('-2147483635_location-id-2', '[KEC-Demo] location-id-2',  NOW(),  NOW(), NULL, '-2147483635', '-2147483635_location-type-id-2', 'location-id-2', '-2147483635_location-id-1', '-2147483635_location-id-1', false, '01FR4M51XJY9E77GSN4QZ1Q8N4/-2147483635_location-id-1/-2147483635_location-id-2'),
  ('-2147483635_location-id-3', '[KEC-Demo] location-id-3',  NOW(),  NOW(), NULL, '-2147483635', '-2147483635_location-type-id-2', 'location-id-3', '-2147483635_location-id-1', '-2147483635_location-id-1', false, '01FR4M51XJY9E77GSN4QZ1Q8N4/-2147483635_location-id-1/-2147483635_location-id-3'),
  ('-2147483635_location-id-4', '[KEC-Demo] location-id-4',  NOW(),  NOW(), NULL, '-2147483635', '-2147483635_location-type-id-2', 'location-id-4', '-2147483635_location-id-1', '-2147483635_location-id-1', false, '01FR4M51XJY9E77GSN4QZ1Q8N4/-2147483635_location-id-1/-2147483635_location-id-4')
  ON CONFLICT DO NOTHING;


\connect mastermgmt;

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('高校3年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483644', '01GB7EAYGAS312J0J2W7H0JR12', 'id_12', NULL, 12)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483644';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('小学4年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483644', '01GB7EAYGAS312J0J2W87JR4N2', 'id_4', NULL, 4)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483644';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('小学3年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483644', '01GB7EAYGAS312J0J2W5APGQYX', 'id_3', NULL, 3)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483644';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('中学1年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483644', '01GB7EAYGAS312J0J2VT50BC67', 'id_7', NULL, 7)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483644';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('中学2年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483644', '01GB7EAYGAS312J0J2W5CRR9RH', 'id_8', NULL, 8)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483644';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('中学3年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483644', '01GB7EAYGAS312J0J2VV98FJSB', 'id_9', NULL, 9)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483644';

INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.live_lesson.enable_live_lesson', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.live_lesson.cloud_record', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.lessonmgmt.zoom_selection', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.lessonmgmt.allow_write_lesson', 'boolean', 'true', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('user.enrollment.update_status_manual', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('hcm.timesheet_management', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('syllabus.learning_material.content_lo', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('payment.order.enable_order_manager', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.assigned_student_list', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.lesson_report.enable_lesson_report', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('user.authentication.ip_address_restriction', 'string', 'off', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('user.authentication.allowed_ip_address', 'string', '', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('syllabus.approve_grading', 'string', 'off', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('general.app_theme', 'string', '#566d8f', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:33:44.623', '2022-12-20 17:33:44.623', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('user.authentication.allowed_address_list', 'string', '', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:33:44.623', '2022-12-20 17:33:44.623', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('general.logo', 'string', 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQbWzBnqkpzJri-KyLdivc6ps9U3NP8XFTpOK-F4sKzAA&s', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:33:44.623', '2022-12-20 17:33:44.623', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.lessonmgmt.lesson_selection', 'string', 'off', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:33:44.623', '2022-12-20 17:33:44.623', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.zoom.config', 'json', '', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:33:44.623', '2022-12-20 17:33:44.623', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.zoom.is_enabled', 'boolean', 'false', 'CONFIGURATION_TYPE_EXTERNAL', '2022-12-20 17:33:44.623', '2022-12-20 17:33:44.623', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('urls_widget', 'json', '', 'CONFIGURATION_TYPE_INTERNAL', '2023-01-05 12:44:31.147', '2023-01-05 12:44:31.147', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('user.student_course.allow_input_student_course', 'string', 'on', 'CONFIGURATION_TYPE_INTERNAL', '2022-12-20 17:09:48.290', '2022-12-20 17:09:48.290', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.live_lesson.view_recording_status', 'boolean', 'false', 'CONFIGURATION_TYPE_INTERNAL', '2023-01-10 07:11:18.200', '2023-01-10 07:11:18.200', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.live_lesson.start_recording_notification', 'boolean', 'false', 'CONFIGURATION_TYPE_INTERNAL', '2023-01-10 07:11:18.200', '2023-01-10 07:11:18.200', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('invoice.invoicemgmt.enable_invoice_manager', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2023-01-10 16:02:07.225', '2023-01-10 16:02:07.225', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('entryexit.entryexitmgmt.enable_entryexit_manager', 'string', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2023-01-12 11:35:11.167', '2023-01-12 11:35:11.167', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('syllabus.grade_book.is_enabled', 'boolean', 'off', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-13 16:18:19.508', '2023-02-13 16:18:19.508', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('mastermgmt.brightcoveconfig.account_id', 'string', '6064018595001', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-14 11:41:49.048', '2023-02-14 11:41:49.048', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('mastermgmt.brightcoveconfig.client_id', 'string', '7f7d1f2e-9a66-4cf5-8187-95aabd9ccaa8', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-14 11:41:49.048', '2023-02-14 11:41:49.048', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('mastermgmt.brightcoveconfig.profile', 'string', 'Asia-PREMIUM (96-1500)', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-14 11:41:49.048', '2023-02-14 11:41:49.048', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('lesson.lessonmgmt.new_student_tag', 'number', '30', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-14 18:00:25.526', '2023-02-14 18:00:25.526', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('syllabus.grade_book.management', 'boolean', 'false', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-22 16:35:43.009', '2023-02-22 16:35:43.009', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('syllabus.study_plan.learning_history.relay_server_url', 'string', '', 'CONFIGURATION_TYPE_INTERNAL', '2023-02-27 17:08:23.043', '2023-02-27 17:08:23.043', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('communication.chat.enable_student_chat', 'boolean', '', 'CONFIGURATION_TYPE_EXTERNAL', '2023-03-01 13:46:49.309', '2023-03-01 13:46:49.309', NULL);
INSERT INTO public.configuration_key
(config_key, value_type, default_value, configuration_type, created_at, updated_at, deleted_at)
VALUES('communication.chat.enable_parent_chat', 'boolean', '', 'CONFIGURATION_TYPE_EXTERNAL', '2023-03-01 13:46:49.309', '2023-03-01 13:46:49.309', NULL);



INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('cb3f2a26-30d7-4616-b9aa-de1dd9ef5ea8', 'lesson.live_lesson.enable_live_lesson', 'on', 'string', NULL, '2022-11-09 17:45:56.137', '2022-11-09 17:45:56.137', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('fd386e82-9f84-4d28-8f1b-d065b881c9db', 'lesson.live_lesson.cloud_record', 'off', 'string', NULL, '2022-11-09 17:45:56.137', '2022-11-09 17:45:56.137', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('68beec9b-9286-4161-9e1d-7d6b6bfa129f', 'lesson.lessonmgmt.zoom_selection', 'off', 'string', NULL, '2022-11-09 17:45:56.137', '2022-11-09 17:45:56.137', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('1fc736ed-d82b-4158-93ab-583cf271c864', 'lesson.lesson_report.enable_lesson_report', 'on', 'string', NULL, '2022-11-09 17:45:56.137', '2022-11-09 17:45:56.137', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('afdaa0b6-39af-4731-a437-2a4b513dc576', 'user.student_course.allow_input_student_course', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('232d5a3c-c8d3-4d06-ac0e-937caea963b8', 'lesson.lessonmgmt.allow_write_lesson', 'true', 'boolean', NULL, '2022-11-09 17:45:56.137', '2022-11-09 17:45:56.137', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('cf555d27-b8a7-418a-905f-9d2f93108f73', 'hcm.timesheet_management', 'on', 'string', NULL, '2022-11-09 17:50:58.746', '2022-11-09 17:50:58.746', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('52df5103-a5f4-497d-a38f-0f7c9acef658', 'payment.order.enable_order_manager', 'off', 'string', NULL, '2022-11-28 14:36:07.047', '2022-11-28 14:36:07.047', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('a110104e-f319-4164-ae22-282717696e31', 'syllabus.learning_material.content_lo', 'on', 'string', NULL, '2022-11-09 17:50:12.362', '2022-11-09 17:50:12.362', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('7e3bdc1a-13d8-4176-93b0-043179bcd765', 'lesson.assigned_student_list', 'on', 'string', NULL, '2022-12-07 11:37:00.826', '2022-12-07 11:37:00.826', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('d5dac144-6ef6-41ab-827c-a5eb1de41b8f', 'urls_widget', '', 'json', NULL, '2023-01-05 12:44:31.147', '2023-01-05 12:44:31.147', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('9a2e1263-dfe0-476e-854a-0f29d07b8b4e', 'lesson.live_lesson.start_recording_notification', 'false', 'boolean', NULL, '2023-01-10 07:11:18.200', '2023-01-10 07:11:18.200', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('491f4759-a6aa-497e-9822-d5ec7a43af12', 'lesson.live_lesson.view_recording_status', 'true', 'boolean', NULL, '2023-01-10 07:11:18.200', '2023-01-10 07:11:18.200', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) 
VALUES
('0b6f273d-d249-4b31-8d4d-c69f51721b9d', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483644'),
('0b6f273d-d249-4b31-8d4d-c69f51721b35', 'user.enrollment.update_status_manual', 'off', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483635'),
('0b6f273d-d249-4b31-8d4d-c69f51721b48', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483648'),
('0b6f273d-d249-4b31-8d4d-c69f51721b30', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483630'),
('0b6f273d-d249-4b31-8d4d-c69f51721b29', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483629')
ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('cd296872-717d-4676-8345-5f2043d92aa8', 'entryexit.entryexitmgmt.enable_entryexit_manager', 'on', 'string', NULL, '2023-01-12 11:35:11.167', '2023-01-12 11:35:11.167', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('d4beb94b-5d59-43eb-942a-16e716432257', 'invoice.invoicemgmt.enable_invoice_manager', 'on', 'string', NULL, '2023-01-10 16:02:07.225', '2023-01-10 16:02:07.225', NULL, '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique  DO UPDATE SET config_value='on';
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('66e86461-d354-4cc3-8db2-14efdff1f9d6', 'syllabus.grade_book.is_enabled', 'off', 'boolean', NULL, '2023-02-13 16:18:19.508', '2023-02-13 16:18:19.508', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('ee672bf9-d4d4-4ba5-9a9d-5f16024e0aeb', 'mastermgmt.brightcoveconfig.account_id', '6064018595001', 'string', NULL, '2023-02-14 11:41:49.048', '2023-02-14 11:41:49.048', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('f5179f11-6686-42d9-a7a8-c8d727d29619', 'mastermgmt.brightcoveconfig.client_id', '7f7d1f2e-9a66-4cf5-8187-95aabd9ccaa8', 'string', NULL, '2023-02-14 11:41:49.048', '2023-02-14 11:41:49.048', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('5e449782-1028-42c9-96c4-d709ef20f30f', 'mastermgmt.brightcoveconfig.profile', 'Asia-PREMIUM (96-1500)', 'string', NULL, '2023-02-14 11:41:49.048', '2023-02-14 11:41:49.048', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('515defaa-197b-42d3-9ca6-52b1a8fc0af1', 'lesson.lessonmgmt.new_student_tag', '30', 'number', NULL, '2023-02-14 18:00:25.526', '2023-02-14 18:00:25.526', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('e750b703-e06f-41de-9c1d-c91a122d0d27', 'syllabus.grade_book.management', 'true', 'boolean', NULL, '2023-02-22 16:35:43.009', '2023-02-22 16:35:43.009', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) VALUES('5a1f8bff-38ef-4cd0-872e-bc4d9cf0bca5', 'syllabus.study_plan.learning_history.relay_server_url', '', 'string', NULL, '2023-02-27 17:08:23.043', '2023-02-27 17:08:23.043', NULL, '-2147483644') ON CONFLICT DO NOTHING;


INSERT INTO public.location_configuration_value
(location_config_id, config_key, location_id, config_value, config_value_type, created_at, updated_at, deleted_at, resource_path)
VALUES('08427b14-f7b1-4227-aecd-7a0aa9f21027', 'communication.chat.enable_student_chat', '01FR4M51XJY9E77GSN4QZ1Q9N5', 'true', 'boolean', '2023-03-01 13:46:49.309', '2023-03-01 13:46:49.309', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.location_configuration_value
(location_config_id, config_key, location_id, config_value, config_value_type, created_at, updated_at, deleted_at, resource_path)
VALUES('bf170623-c9be-4e88-acf3-fe3ca7737445', 'communication.chat.enable_parent_chat', '01FR4M51XJY9E77GSN4QZ1Q9N5', 'true', 'boolean', '2023-03-01 13:46:49.309', '2023-03-01 13:46:49.309', NULL, '-2147483644') ON CONFLICT DO NOTHING;


-- allow teacher view lesson.report
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GC8P2NSD77DMVPTF4YPM4HFF', 'lesson.report.review', now(), now(), '-2147483644')
  ON CONFLICT DO NOTHING;


INSERT INTO public.permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01GC8P2NSD77DMVPTF4YPM4HFF', '01G1GQEKEHXKSM78NBW96NJ7H8', now(), now(), '-2147483644'),
  ('01GC8P2NSD77DMVPTF4YPM4HFF', '01G7XGB49W2PCQPHNBE6SAZ443', now(), now(), '-2147483644'),

  ('01GC8P2NSD77DMVPTF4YPM4HFF', '01G7XGB49W2PCQPHNBE6SAZ441', now(), now(), '-2147483644'),
  ('01GC8P2NSD77DMVPTF4YPM4HFF', '01G7XGB49W2PCQPHNBE6SAZ442', now(), now(), '-2147483644'),
  ('01GC8P2NSD77DMVPTF4YPM4HFF', '01G7XGB49W2PCQPHNBE6SAZ444', now(), now(), '-2147483644'),
  ('01GC8P2NSD77DMVPTF4YPM4HFF', '01G7XGB49W2PCQPHNBE6SAZ445', now(), now(), '-2147483644')
  ON CONFLICT DO NOTHING;

-- update tenant and project local
update organizations set tenant_id = 'withus-managara-base-0wf23', domain_name = 'managara-base' where resource_path = '-2147483630';
update organizations set tenant_id = 'withus-managara-hs-t5fuk', domain_name= 'managara-hs' where resource_path = '-2147483629';
update organizations set tenant_id = 'manabie-0nl6t', domain_name= 'manabie' where resource_path = '-2147483648';
update organizations set tenant_id = 'end-to-end-dopvo', domain_name= 'e2e' where resource_path = '-2147483644';

INSERT INTO grade (grade_id,"name",is_archived,partner_internal_id,updated_at,created_at,deleted_at,resource_path,"sequence",remarks) VALUES
  ('01GWK2BWC0JSJQYAAHJ7682YWQ','Grade 6',false,'6',now(),now(),NULL,'-2147483629',6,''),
  ('01GWK2BWC0JSJQYAAHJB0KQ97N','Grade 7',false,'7',now(),now(),NULL,'-2147483629',7,''),
  ('01GWK2BWC0JSJQYAAHJDJ9QC0T','Grade 8',false,'8',now(),now(),NULL,'-2147483629',8,''),
  ('01GWK2BWC0JSJQYAAHJHCSHCYX','Grade 9',false,'9',now(),now(),NULL,'-2147483629',9,''),
  ('01GWK2BWC0JSJQYAAHJKJ12SPH','Grade 10',false,'10',now(),now(),NULL,'-2147483629',10,''),
  ('01GWK2B42AY86RN1GSQ25Y35WG','Grade 1',false,'1',now(),now(),NULL,'-2147483630',1,''),
  ('01GWK2B42AY86RN1GSQ4SFKZWS','Grade 2',false,'2',now(),now(),NULL,'-2147483630',2,''),
  ('01GWK2B42AY86RN1GSQ7S15709','Grade 3',false,'3',now(),now(),NULL,'-2147483630',3,''),
  ('01GWK2B42AY86RN1GSQAWJNVMW','Grade 4',false,'4',now(),now(),NULL,'-2147483630',4,''),
  ('01GWK2B42AY86RN1GSQEAER3ZZ','Grade 5',false,'5',now(),now(),NULL,'-2147483630',5,''),
  ('-2147483635_grade_01','[KEC-Demo] Grade 1',false,'grade_01',now(),now(),NULL,'-2147483635',11,''),
  ('-2147483635_grade_02','[KEC-Demo] Grade 2',false,'grade_02',now(),now(),NULL,'-2147483635',12,''),
  ('-2147483635_grade_03','[KEC-Demo] Grade 3',false,'grade_03',now(),now(),NULL,'-2147483635',13,''),
  ('-2147483635_grade_04','[KEC-Demo] Grade 4',false,'grade_04',now(),now(),NULL,'-2147483635',14,''),
  ('-2147483635_grade_05','[KEC-Demo] Grade 5',false,'grade_05',now(),now(),NULL,'-2147483635',15,''),
  ('-2147483648_grade_01','[Manabie] Grade 1',false,'grade_01',now(),now(),NULL,'-2147483648',16,''),
  ('-2147483648_grade_02','[Manabie] Grade 2',false,'grade_02',now(),now(),NULL,'-2147483648',17,''),
  ('-2147483648_grade_03','[Manabie] Grade 3',false,'grade_03',now(),now(),NULL,'-2147483648',18,''),
  ('-2147483648_grade_04','[Manabie] Grade 4',false,'grade_04',now(),now(),NULL,'-2147483648',19,''),
  ('-2147483648_grade_05','[Manabie] Grade 5',false,'grade_05',now(),now(),NULL,'-2147483648',20,'')
  ON CONFLICT DO NOTHING;

INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) 
VALUES
('9e2052ce-10fe-46b8-8568-be119457971d', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483644'),
('ef049623-1b7c-47de-9ed0-63c9883020e3', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483648'),
('40f6fcdb-11d0-42ea-8682-5cd82591d445', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483630'),
('196b385b-db80-4d91-af1a-f17965931e18', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483629')
ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO UPDATE SET config_value = 'on';

INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) 
VALUES
('95dd703f-e7cb-4fd4-9b59-2b44e598946e', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483644'),
('b62595ee-6107-4ea5-bcf4-b19fd68efaf8', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483648'),
('01bb3183-bb2e-4959-a7a7-eccc738a6f05', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483630'),
('4fc12054-6b60-4abc-b6ea-8772468a5779', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483629'),
('d073fe4d-d3db-45d2-8490-0057f8a09319', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483635')
ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO UPDATE SET config_value = 'on';