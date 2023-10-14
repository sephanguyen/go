package form_partner

var (
	mapFieldRenseikaiAndBestcoPartner = map[string][]string{
		"remark_internal": {"remarks"},
		"lesson_content":  {"lesson_content", "lesson_homework"},
	}
)

type ReseikaiAndBestcoPartner struct {
	*FormPartner
}

func (p *ReseikaiAndBestcoPartner) getMapNewField() map[string][]string {
	return mapFieldRenseikaiAndBestcoPartner
}

func (p *ReseikaiAndBestcoPartner) GetConfigFormName() string {
	return "FORM_CONFIG_LESSON_REPORT_IND_UPDATE"
}
func (p *ReseikaiAndBestcoPartner) GetConfigForm() string {
	return `{
		"sections": [
			{
				"fields": [
					{
						"label": {
							"i18n": {
								"translations": {
									"en": "Attendance",
									"ja": "出席情報",
									"vi": "Attendance"
								},
								"fallback_language": "ja"
							}
						},
						"field_id": "attendance_label",
						"value_type": "VALUE_TYPE_NULL",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 12,
								"xs": 12
							}
						},
						"component_props": {
							"variant": "body2"
						},
						"component_config": {
							"type": "TYPOGRAPHY"
						}
					},
					{
						"field_id": "attendance_status",
						"value_type": "VALUE_TYPE_STRING",
						"is_internal": false,
						"is_required": true,
						"display_config": {
							"grid_size": {
								"md": 6,
								"xs": 6
							}
						},
						"component_config": {
							"type": "ATTENDANCE_STATUS"
						}
					},
					{
						"field_id": "attendance_notice",
						"value_type": "VALUE_TYPE_STRING",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 6,
								"xs": 6
							}
						},
						"component_config": {
							"type": "ATTENDANCE_NOTICE"
						}
					},
					{
						"field_id": "attendance_reason",
						"value_type": "VALUE_TYPE_STRING",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 6,
								"xs": 6
							}
						},
						"component_config": {
							"type": "ATTENDANCE_REASON"
						}
					},
					{
						"field_id": "attendance_note",
						"value_type": "VALUE_TYPE_STRING",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 6,
								"xs": 6
							}
						},
						"component_config": {
							"type": "ATTENDANCE_NOTE"
						}
					}
				],
				"section_id": "attendance_section_id",
				"section_name": "attendance"
			},
			{
				"fields": [
					{
						"label": {
							"i18n": {
								"translations": {
									"en": "This Lesson",
									"ja": "今回の授業",
									"vi": "This Lesson"
								},
								"fallback_language": "ja"
							}
						},
						"field_id": "this_lesson_label",
						"value_type": "VALUE_TYPE_NULL",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 9,
								"xs": 9
							}
						},
						"component_props": {
							"variant": "body2"
						},
						"component_config": {
							"type": "TYPOGRAPHY"
						}
					},
					{
						"field_id": "lesson_previous_report_action",
						"value_type": "VALUE_TYPE_NULL",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 3,
								"xs": 3
							}
						},
						"component_config": {
							"type": "BUTTON_PREVIOUS_REPORT"
						}
					},
					{
						"label": {
							"i18n": {
								"translations": {
									"en": "Content",
									"ja": "授業内容",
									"vi": "Content"
								},
								"fallback_language": "ja"
							}
						},
						"field_id": "lesson_content",
						"value_type": "VALUE_TYPE_STRING",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 12,
								"xs": 12
							}
						},
						"component_props": {
							"InputProps": {
								"rows": 6,
								"multiline": true
							}
						},
						"component_config": {
							"type": "TEXT_FIELD_AREA"
						}
					}
				],
				"section_id": "this_lesson_section_id",
				"section_name": "this_lesson"
			},
			{
				"fields": [
					{
						"label": {
							"i18n": {
								"translations": {
									"en": "Remarks",
									"ja": "特筆事項",
									"vi": "Remarks"
								},
								"fallback_language": "ja"
							}
						},
						"field_id": "remarks_label",
						"value_type": "VALUE_TYPE_NULL",
						"is_internal": false,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 12,
								"xs": 12
							}
						},
						"component_props": {
							"variant": "body2"
						},
						"component_config": {
							"type": "TYPOGRAPHY"
						}
					},
					{
						"label": {
							"i18n": {
								"translations": {
									"en": "Remark",
									"ja": "特筆事項",
									"vi": "Remark"
								},
								"fallback_language": "ja"
							}
						},
						"field_id": "remark_internal",
						"value_type": "VALUE_TYPE_STRING",
						"is_internal": true,
						"is_required": false,
						"display_config": {
							"grid_size": {
								"md": 12,
								"xs": 12
							}
						},
						"component_props": {
							"InputProps": {
								"rows": 6,
								"multiline": true
							}
						},
						"component_config": {
							"type": "TEXT_FIELD_AREA",
							"question_mark": {
								"message": {
									"i18n": {
										"translations": {
											"en": "This is an internal memo, it will not be shared with parents",
											"ja": "これは社内用メモです。保護者には共有されません",
											"vi": "This is an internal memo, it will not be shared with parents"
										},
										"fallback_language": "ja"
									}
								}
							}
						}
					}
				],
				"section_id": "remark_section_id",
				"section_name": "remark"
			}
		]
	}`
}
