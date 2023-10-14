package lessonmgmt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/proto"
)

func (s *Suite) AFormConfigForFeature(ctx context.Context, feature string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CurrentSchoolID == 0 || stepState.CurrentSchoolID == 1 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	var (
		featureName string
		jsonConfig  pgtype.JSONB
	)
	switch feature {
	case "individual lesson report":
		featureName = string(lesson_report_consts.FeatureNameIndividualUpdateLessonReport)
		jsonConfig = database.JSONB(`
		{
			"sections": [
				{
					"section_id": "section-id-0",
					"section_name": "section-name",
					"fields": [
						{
							"field_id": "attendance_status",
							"label": "attendance status",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						},
						{
							"field_id": "attendance_remark",
							"label": "attendance remark",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						}
					]
				},
				{
					"section_id": "section-id-1",
					"section_name": "section-name",
					"fields": [
						{
							"field_id": "ordinal-number",
							"label": "display name 1",
							"value_type": "VALUE_TYPE_INT",
							"is_required": false,
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						},
						{
							"field_id": "title",
							"label": "display name 2",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						}
					]
				},
				{
					"section_id": "section-id-2",
					"section_name": "section-name-2",
					"fields": [
						{
							"field_id": "buddy",
							"label": "display name 3",
							"value_type": "VALUE_TYPE_STRING_SET",
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						},
						{
							"field_id": "scores",
							"label": "display name 4",
							"value_type": "VALUE_TYPE_INT_ARRAY",
							"is_required": false,
							"component_props": {}
						},
						{
							"field_id": "comments",
							"label": "display name 4",
							"value_type": "VALUE_TYPE_STRING_ARRAY",
							"is_required": false
						},
						{
							"field_id": "finished-exams",
							"label": "display name 4",
							"value_type": "VALUE_TYPE_INT_SET",
							"is_required": false
						},
						{
							"field_id": "is-pass-lesson",
							"label": "display name 2",
							"value_type": "VALUE_TYPE_BOOL",
							"is_required": false,
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						}
					]
				}
			]
		}
	`)
	case "group lesson report":
		featureName = string(lesson_report_consts.FeatureNameGroupLessonReport)
		jsonConfig = database.JSONB(`
		{
			"sections": [{
					"section_id": "this_lesson_id",
					"section_name": "this_lesson",
					"fields": [{
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
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 10,
									"xs": 10
								}
							},
							"component_props": {
								"variant": "h6"
							},
							"component_config": {
								"type": "TYPOGRAPHY"
							}
						},
						{
							"field_id": "lesson_previous_report_action",
							"value_type": "VALUE_TYPE_NULL",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 2,
									"xs": 2
								}
							},
							"component_config": {
								"type": "BUTTON_PREVIOUS_REPORT"
							}
						},
						{
							"field_id": "is-pass-lesson",
							"label": "display name 2",
							"value_type": "VALUE_TYPE_BOOL",
							"is_required": false,
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						},
						{
							"field_id": "scores",
							"label": "display name 4",
							"value_type": "VALUE_TYPE_INT_ARRAY",
							"is_required": false,
							"component_props": {}
						},
						{
							"field_id": "buddy",
							"label": "display name 3",
							"value_type": "VALUE_TYPE_STRING_SET",
							"component_props": {},
							"component_config": {
								"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
							},
							"display_config": {
								"is_label": true,
								"size": {}
							}
						},
						{
							"field_id": "finished-exams",
							"label": "display name 4",
							"value_type": "VALUE_TYPE_INT_SET",
							"is_required": false
						},
						{
							"field_id": "comments",
							"label": "display name 4",
							"value_type": "VALUE_TYPE_STRING_ARRAY",
							"is_required": false
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "Content",
										"ja": "備考",
										"vi": "Content"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "content",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": true,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 6,
									"xs": 6
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
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "Remark",
										"ja": "備考",
										"vi": "Remark"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "lesson_remark",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"is_internal": true,
							"display_config": {
								"grid_size": {
									"md": 6,
									"xs": 6
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
					]
				},
				{
					"section_id": "next_lesson_id",
					"section_name": "next_lesson",
					"fields": [{
							"label": {
								"i18n": {
									"translations": {
										"en": "Next Lesson",
										"ja": "次回の授業",
										"vi": "Next Lesson"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "next_lesson_label",
							"value_type": "VALUE_TYPE_NULL",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 12,
									"xs": 12
								}
							},
							"component_props": {
								"variant": "h6"
							},
							"component_config": {
								"type": "TYPOGRAPHY"
							}
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "Homework",
										"ja": "備考",
										"vi": "Homework"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "homework",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": true,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 6,
									"xs": 6
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
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "Announcement",
										"ja": "お知らせ",
										"vi": "Announcement"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "announcement",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 6,
									"xs": 6
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
					]
				},
				{
					"section_id": "student_list_id",
					"section_name": "student_list",
					"fields": [{
							"label": {
								"i18n": {
									"translations": {
										"en": "Student List",
										"ja": "出席情報",
										"vi": "Student List"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "student_list_label",
							"value_type": "VALUE_TYPE_NULL",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 12,
									"xs": 12
								}
							},
							"component_props": {
								"variant": "h6"
							},
							"component_config": {
								"type": "TYPOGRAPHY"
							}
						},
						{
							"field_id": "student_list_tables",
							"value_type": "VALUE_TYPE_NULL",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"grid_size": {
									"md": 12,
									"xs": 12
								}
							},
							"component_props": {
								"toggleButtons": [{
										"label": {
											"i18n": {
												"translations": {
													"en": "Performance",
													"ja": "成績",
													"vi": "Performance"
												},
												"fallback_language": "ja"
											}
										},
										"field_id": "performance"
									},
									{
										"label": {
											"i18n": {
												"translations": {
													"en": "Remark",
													"ja": "備考",
													"vi": "Remark"
												},
												"fallback_language": "ja"
											}
										},
										"field_id": "remark"
									}
								]
							},
							"component_config": {
								"type": "TOGGLE_TABLE"
							}
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "Homework Completion",
										"ja": "宿題提出",
										"vi": "Homework Completion"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "homework_completion",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"table_size": {
									"width": "22%"
								}
							},
							"component_props": {
								"options": [{
										"key": "COMPLETED",
										"icon": "CircleOutlined"
									},
									{
										"key": "IN_PROGRESS",
										"icon": "ChangeHistoryOutlined"
									},
									{
										"key": "INCOMPLETE",
										"icon": "CloseOutlined"
									}
								],
								"optionIconLabelKey": "icon",
								"valueKey": "key",
								"placeholder": {
									"i18n": {
										"translations": {
											"en": "Homework Completion",
											"ja": "宿題提出",
											"vi": "Homework Completion"
										},
										"fallback_language": "ja"
									}
								}
							},
							"component_config": {
								"type": "SELECT_ICON",
								"table_key": "performance",
								"has_bulk_action": true
							}
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "in-lesson Quiz",
										"ja": "小テスト",
										"vi": "in-lesson Quiz"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "in_lesson_quiz",
							"value_type": "VALUE_TYPE_INT",
							"is_required": false,
							"is_internal": false,
							"display_config": {
								"table_size": {
									"width": "22%"
								}
							},
							"component_props": {
								"placeholder": {
									"i18n": {
										"translations": {
											"en": "in-lesson Quiz",
											"ja": "小テスト",
											"vi": "in-lesson Quiz"
										},
										"fallback_language": "ja"
									}
								}
							},
							"component_config": {
								"type": "TEXT_FIELD_PERCENTAGE",
								"table_key": "performance",
								"has_bulk_action": true
							}
						},
						{
							"label": {
								"i18n": {
									"translations": {
										"en": "Remark",
										"ja": "提出状況",
										"vi": "Remark"
									},
									"fallback_language": "ja"
								}
							},
							"field_id": "student_remark",
							"value_type": "VALUE_TYPE_STRING",
							"is_required": false,
							"is_internal": true,
							"display_config": {
								"table_size": {
									"width": "70%"
								}
							},
							"component_props": {
								"placeholder": {
									"i18n": {
										"translations": {
											"en": "Remark",
											"ja": "提出状況",
											"vi": "Remark"
										},
										"fallback_language": "ja"
									}
								}
							},
							"component_config": {
								"type": "TEXT_FIELD",
								"table_key": "remark",
								"has_bulk_action": false,
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
					]
				}
			]
		}
		`)
	}

	stepState.FormConfigID = idutil.ULIDNow()
	now := time.Now()

	config := &repo.PartnerFormConfigDTO{
		FormConfigID: database.Text(stepState.FormConfigID),
		PartnerID:    database.Int4(stepState.CurrentSchoolID),
		CreatedAt:    database.Timestamptz(now),
		UpdatedAt:    database.Timestamptz(now),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
		FeatureName:    database.Text(featureName),
		FormConfigData: jsonConfig,
	}

	fieldNames, args := config.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		config.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := s.BobDB.Exec(ctx, query, args...); err != nil {
		return ctx, fmt.Errorf("insert a form config: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserSubmitANewLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				AttendanceRemark: "very good",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.BobConn).SubmitLessonReport(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
		stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
			stepState.StudentIds[0]: {
				{
					FieldID:          database.Text("ordinal-number"),
					IntValue:         database.Int4(5),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("title"),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
					StringValue:      database.Text("monitor"),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("scores"),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("comments"),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
				},
				{
					FieldID:        database.Text("buddy"),
					ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
				},
				{
					FieldID:     database.Text("finished-exams"),
					ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
				},
			},
			stepState.StudentIds[1]: {
				{
					FieldID:          database.Text("ordinal-number"),
					IntValue:         database.Int4(15),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
			},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserSaveDraftLessonReportGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.WriteGroupLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*lpb.GroupLessonReportDetails{
			{
				StudentId: stepState.StudentIds[0],
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "content",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "content",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "lesson_remark",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "lesson_remark",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "homework",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "homework",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "in_lesson_quiz",
						Value: &lpb.DynamicFieldValue_IntValue{
							IntValue: 0,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &lpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &lpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &lpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      cpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &lpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6},
							},
						},
					},
				},
			},
			{
				StudentId: stepState.StudentIds[1],
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "content",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "content 1",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "lesson_remark",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "lesson_remark 1",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	if len(stepState.LessonReportID) > 0 {
		req.LessonReportId = stepState.LessonReportID
	}
	stepState.Request = req
	res, err := lpb.NewLessonReportModifierServiceClient(s.LessonMgmtConn).SaveDraftGroupLessonReport(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
		stepState.LessonDynamicFieldValuesByUser = map[string][]*repo.PartnerDynamicFormFieldValueDTO{
			stepState.StudentIds[0]: {
				{
					FieldID:          database.Text("content"),
					StringValue:      database.Text("content"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("lesson_remark"),
					StringValue:      database.Text("lesson_remark"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("homework"),
					StringValue:      database.Text("homework"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("in_lesson_quiz"),
					IntValue:         database.Int4(0),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("comments"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
				},
				{
					FieldID:        database.Text("buddy"),
					ValueType:      database.Text(cpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
				},
				{
					FieldID:     database.Text("finished-exams"),
					ValueType:   database.Text(cpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
				},
			},
			stepState.StudentIds[1]: {
				{
					FieldID:          database.Text("content"),
					StringValue:      database.Text("content 1"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("lesson_remark"),
					StringValue:      database.Text("lesson_remark 1"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
			},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserSaveDraftLessonReportIndividual(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.WriteIndividualLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*lpb.IndividualLessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &lpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &lpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &lpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &lpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &lpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      cpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &lpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &lpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &lpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	req.Details[0].AttendanceStatus = lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY
	if len(stepState.LessonReportID) > 0 {
		req.LessonReportId = stepState.LessonReportID
	}
	stepState.Request = req
	res, err := lpb.NewLessonReportModifierServiceClient(s.LessonMgmtConn).SaveDraftIndividualLessonReport(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
		stepState.LessonDynamicFieldValuesByUser = map[string][]*repo.PartnerDynamicFormFieldValueDTO{
			stepState.StudentIds[0]: {
				{
					FieldID:          database.Text("ordinal-number"),
					IntValue:         database.Int4(5),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("title"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					StringValue:      database.Text("monitor"),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("scores"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("comments"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
				},
				{
					FieldID:        database.Text("buddy"),
					ValueType:      database.Text(cpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
				},
				{
					FieldID:     database.Text("finished-exams"),
					ValueType:   database.Text(cpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
				},
			},
			stepState.StudentIds[1]: {
				{
					FieldID:          database.Text("ordinal-number"),
					IntValue:         database.Int4(15),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
			},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) UserSubmitLessonReportGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.WriteGroupLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*lpb.GroupLessonReportDetails{
			{
				StudentId: stepState.StudentIds[0],
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "content",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "content",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "lesson_remark",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "lesson_remark",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "homework",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "homework",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "in_lesson_quiz",
						Value: &lpb.DynamicFieldValue_IntValue{
							IntValue: 0,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &lpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &lpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &lpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      cpb.ValueType_VALUE_TYPE_INT_ARRAY,
						Value: &lpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6},
							},
						},
					},
				},
			},
			{
				StudentId: stepState.StudentIds[1],
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "content",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "content 1",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "lesson_remark",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "lesson_remark 1",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "homework",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "homework",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	if len(stepState.LessonReportID) > 0 {
		req.LessonReportId = stepState.LessonReportID
	}
	ctx, err := s.subscribeLessonUpdatedForUpdateSchedulingStatus(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = req
	res, err := lpb.NewLessonReportModifierServiceClient(s.LessonMgmtConn).SubmitGroupLessonReport(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
		stepState.LessonDynamicFieldValuesByUser = map[string][]*repo.PartnerDynamicFormFieldValueDTO{
			stepState.StudentIds[0]: {
				{
					FieldID:          database.Text("content"),
					StringValue:      database.Text("content"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("lesson_remark"),
					StringValue:      database.Text("lesson_remark"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("homework"),
					StringValue:      database.Text("homework"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("in_lesson_quiz"),
					IntValue:         database.Int4(0),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("comments"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
				},
				{
					FieldID:        database.Text("buddy"),
					ValueType:      database.Text(cpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
				},
				{
					FieldID:     database.Text("finished-exams"),
					ValueType:   database.Text(cpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
				},
			},
			stepState.StudentIds[1]: {
				{
					FieldID:          database.Text("content"),
					StringValue:      database.Text("content 1"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("lesson_remark"),
					StringValue:      database.Text("lesson_remark 1"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("homework"),
					StringValue:      database.Text("homework"),
					ValueType:        database.Text(cpb.ValueType_VALUE_TYPE_STRING.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
			},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) UserSubmitLessonReportIndividual(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.WriteIndividualLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*lpb.IndividualLessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				AttendanceRemark: "very good",
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &lpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &lpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &lpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &lpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &lpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      cpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &lpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      cpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &lpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*lpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &lpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &lpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        cpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	if len(stepState.LessonReportID) > 0 {
		req.LessonReportId = stepState.LessonReportID
	}
	ctx, err := s.subscribeLessonUpdatedForUpdateSchedulingStatus(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = req
	res, err := lpb.NewLessonReportModifierServiceClient(s.LessonMgmtConn).SubmitIndividualLessonReport(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
		stepState.LessonDynamicFieldValuesByUser = map[string][]*repo.PartnerDynamicFormFieldValueDTO{
			stepState.StudentIds[0]: {
				{
					FieldID:          database.Text("ordinal-number"),
					IntValue:         database.Int4(5),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("title"),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
					StringValue:      database.Text("monitor"),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("scores"),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("comments"),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
				},
				{
					FieldID:        database.Text("buddy"),
					ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
					StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
				},
				{
					FieldID:     database.Text("finished-exams"),
					ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
					IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
				},
			},
			stepState.StudentIds[1]: {
				{
					FieldID:          database.Text("ordinal-number"),
					IntValue:         database.Int4(15),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
				{
					FieldID:          database.Text("is-pass-lesson"),
					BoolValue:        database.Bool(true),
					ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
					FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
				},
			},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) subscribeLessonUpdatedForUpdateSchedulingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonUpdatedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		stepState.FoundChanForJetStream <- r.Message
		return true, nil
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonUpdatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) mustHaveEventFromStatusToAfterStatus(ctx context.Context, mustHaveEvent, statusString, newStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if mustHaveEvent == "yes" {
		timer := time.NewTimer(time.Minute * 1)
		allowedSchedulingStatuses := make(map[lesson_domain.LessonSchedulingStatus]bool)
		allowedSchedulingStatuses[lesson_domain.LessonSchedulingStatus(newStatus)] = false
		if len(statusString) > 0 {
			statuses := strings.Split(statusString, ",")
			for _, s := range statuses {
				allowedSchedulingStatuses[lesson_domain.LessonSchedulingStatus(s)] = false
			}
		}
		i := 0
		for {
			select {
			case data := <-stepState.FoundChanForJetStream:
				switch v := data.(type) {
				case *bpb.EvtLesson_UpdateLesson_:
					if v.UpdateLesson.LessonId == stepState.CurrentLessonID {
						status := lesson_domain.LessonSchedulingStatus(v.UpdateLesson.GetSchedulingStatusAfter().String())
						if !allowedSchedulingStatuses[status] {
							allowedSchedulingStatuses[status] = true
							i++
							if i == len(allowedSchedulingStatuses) {
								return StepStateToContext(ctx, stepState), nil
							}
						}
					}
				}

			case <-ctx.Done():
				return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) LessonmgmtHaveANewDraftLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch stepState.CurrentTeachingMethod {
	case "individual":
		return s.checkLessonReportIndividual(ctx, lesson_report_consts.ReportSubmittingStatusSaved)
	case "group":
		return s.checkLessonReportGroup(ctx, lesson_report_consts.ReportSubmittingStatusSaved)
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("invalid lesson report state")
}

func (s *Suite) LessonmgmtHaveANewSubmittedLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch stepState.CurrentTeachingMethod {
	case "individual":
		return s.checkLessonReportIndividual(ctx, lesson_report_consts.ReportSubmittingStatusSubmitted)
	case "group":
		return s.checkLessonReportGroup(ctx, lesson_report_consts.ReportSubmittingStatusSubmitted)
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("invalid lesson report state")
}

func (s *Suite) checkLessonReportIndividual(ctx context.Context, status lesson_report_consts.ReportSubmittingStatus) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*lpb.WriteIndividualLessonReportRequest)

	// get lesson report record
	fields, _ := (&repo.LessonReportDTO{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_reports
		WHERE lesson_report_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	report := repo.LessonReportDTO{}
	err := database.Select(ctx, s.BobDBTrace, query, &stepState.LessonReportID).ScanOne(&report)
	if err != nil {
		return ctx, fmt.Errorf("database.Select: %w", err)
	}

	if report.ReportSubmittingStatus.String != string(status) {
		return ctx, fmt.Errorf("expected report_submitting_status %s but got %s", status, report.ReportSubmittingStatus.String)
	}
	if report.LessonID.String != req.LessonId {
		return ctx, fmt.Errorf("expected lesson_id %s but got %s", req.LessonId, report.LessonID.String)
	}
	if len(report.FormConfigID.String) == 0 {
		return ctx, fmt.Errorf("expected form_config_id but got empty")
	}
	if report.CreatedAt.Status != pgtype.Present {
		return ctx, fmt.Errorf("expected created_at field but got empty")
	}
	if report.UpdatedAt.Status != pgtype.Present {
		return ctx, fmt.Errorf("expected updated_at field but got empty")
	}

	// get lesson details
	details, err := (&repo.LessonReportDetailRepo{}).GetByLessonReportID(ctx, s.BobDBTrace, stepState.LessonReportID)
	if err != nil {
		return ctx, err
	}
	detailIDs := make([]string, 0, len(details))
	for i := range req.Details {
		detailIDs = append(detailIDs, details[i].LessonReportDetailID)
	}
	if err = s.checkLessonReportIndividualDetails(req.Details, details); err != nil {
		return ctx, err
	}

	// get detail's field values
	values, err := s.getPartnerFormDynamicFieldValues(ctx, detailIDs)
	if err != nil {
		return ctx, err
	}
	valuesByDetailID := make(map[string]repo.PartnerDynamicFormFieldValueDTOs)
	for _, value := range values {
		v := valuesByDetailID[value.LessonReportDetailID.String]
		v = append(v, value)
		valuesByDetailID[value.LessonReportDetailID.String] = v
	}
	valuesByUserID := make(map[string]repo.PartnerDynamicFormFieldValueDTOs)
	for _, detail := range details {
		valuesByUserID[detail.StudentID] = valuesByDetailID[detail.LessonReportDetailID]
	}
	if err = s.checkLessonReportDetailFieldValues(valuesByUserID, stepState.LessonDynamicFieldValuesByUser); err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkLessonReportGroup(ctx context.Context, status lesson_report_consts.ReportSubmittingStatus) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*lpb.WriteGroupLessonReportRequest)

	// get lesson report record
	fields, _ := (&repo.LessonReportDTO{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_reports
		WHERE lesson_report_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	report := repo.LessonReportDTO{}
	err := database.Select(ctx, s.BobDBTrace, query, &stepState.LessonReportID).ScanOne(&report)
	if err != nil {
		return ctx, fmt.Errorf("database.Select: %w", err)
	}

	if report.ReportSubmittingStatus.String != string(status) {
		return ctx, fmt.Errorf("expected report_submitting_status %s but got %s", status, report.ReportSubmittingStatus.String)
	}
	if report.LessonID.String != req.LessonId {
		return ctx, fmt.Errorf("expected lesson_id %s but got %s", req.LessonId, report.LessonID.String)
	}
	if len(report.FormConfigID.String) == 0 {
		return ctx, fmt.Errorf("expected form_config_id but got empty")
	}
	if report.CreatedAt.Status != pgtype.Present || report.CreatedAt.Time.IsZero() {
		return ctx, fmt.Errorf("expected created_at field but got empty")
	}
	if report.UpdatedAt.Status != pgtype.Present || report.UpdatedAt.Time.IsZero() {
		return ctx, fmt.Errorf("expected updated_at field but got empty")
	}

	// get lesson details
	details, err := (&repo.LessonReportDetailRepo{}).GetByLessonReportID(ctx, s.BobDBTrace, stepState.LessonReportID)
	if err != nil {
		return ctx, err
	}
	detailIDs := make([]string, 0, len(details))
	for i := range req.Details {
		detailIDs = append(detailIDs, details[i].LessonReportDetailID)
	}
	if err = s.checkLessonReportGroupDetails(req.Details, details); err != nil {
		return ctx, err
	}

	// get detail's field values
	values, err := s.getPartnerFormDynamicFieldValues(ctx, detailIDs)
	if err != nil {
		return ctx, err
	}
	valuesByDetailID := make(map[string]repo.PartnerDynamicFormFieldValueDTOs)
	for _, value := range values {
		v := valuesByDetailID[value.LessonReportDetailID.String]
		v = append(v, value)
		valuesByDetailID[value.LessonReportDetailID.String] = v
	}
	valuesByUserID := make(map[string]repo.PartnerDynamicFormFieldValueDTOs)
	for _, detail := range details {
		valuesByUserID[detail.StudentID] = valuesByDetailID[detail.LessonReportDetailID]
	}
	if err = s.checkLessonReportDetailFieldValues(valuesByUserID, stepState.LessonDynamicFieldValuesByUser); err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkLessonReportIndividualDetails(req []*lpb.IndividualLessonReportDetail, actual domain.LessonReportDetails) error {
	for i, expected := range req {
		if actual[i].StudentID != expected.StudentId {
			return fmt.Errorf("expected student_id %s but got %s", expected.StudentId, actual[i].StudentID)
		}
		if actual[i].CreatedAt.IsZero() {
			return fmt.Errorf("expected created_at field but got empty")
		}
		if actual[i].UpdatedAt.IsZero() {
			return fmt.Errorf("expected updated_at field but got empty")
		}
	}
	return nil
}

func (s *Suite) checkLessonReportGroupDetails(req []*lpb.GroupLessonReportDetails, actual domain.LessonReportDetails) error {
	for i, expected := range req {
		if actual[i].StudentID != expected.StudentId {
			return fmt.Errorf("expected student_id %s but got %s", expected.StudentId, actual[i].StudentID)
		}
		if actual[i].CreatedAt.IsZero() {
			return fmt.Errorf("expected created_at field but got empty")
		}
		if actual[i].UpdatedAt.IsZero() {
			return fmt.Errorf("expected updated_at field but got empty")
		}
	}
	return nil
}

func (s *Suite) getPartnerFormDynamicFieldValues(ctx context.Context, lessonReportDetailIDs []string) (repo.PartnerDynamicFormFieldValueDTOs, error) {
	fields, _ := (&repo.PartnerDynamicFormFieldValueDTO{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM partner_dynamic_form_field_values
		WHERE lesson_report_detail_id = ANY($1) AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	values := repo.PartnerDynamicFormFieldValueDTOs{}
	err := database.Select(ctx, s.BobDBTrace, query, &lessonReportDetailIDs).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return values, nil
}

func (s *Suite) checkLessonReportDetailFieldValues(valuesByUserID map[string]repo.PartnerDynamicFormFieldValueDTOs, dynamicFieldValuesByUser map[string][]*repo.PartnerDynamicFormFieldValueDTO) error {
	for userID, actualValues := range valuesByUserID {
		expected, ok := dynamicFieldValuesByUser[userID]
		if !ok {
			return fmt.Errorf("no fields values of lesson report details which have user id %s", userID)
		}

		expectedByFieldID := make(map[string]*repo.PartnerDynamicFormFieldValueDTO)
		for i := range expected {
			expectedByFieldID[expected[i].FieldID.String] = expected[i]
		}

		if len(expectedByFieldID) != len(actualValues) {
			return fmt.Errorf("expected %d values but got %d", len(expectedByFieldID), len(actualValues))
		}

		for _, actualValue := range actualValues {
			v, ok := expectedByFieldID[actualValue.FieldID.String]
			if !ok {
				return fmt.Errorf("expected field %s's values but got null", actualValue.FieldID.String)
			}

			if string(v.FieldRenderGuide.Bytes) != string(actualValue.FieldRenderGuide.Bytes) {
				return fmt.Errorf("expected field render %s guide but got %s", string(v.FieldRenderGuide.Bytes), string(actualValue.FieldRenderGuide.Bytes))
			}
			if v.IntValue.Status == pgtype.Present && (v.IntValue.Int != actualValue.IntValue.Int) {
				return fmt.Errorf("expected int value %d but got %d", v.IntValue.Int, actualValue.IntValue.Int)
			}
			if v.StringValue.Status == pgtype.Present && (v.StringValue.String != actualValue.StringValue.String) {
				return fmt.Errorf("expected string value %s but got %s", v.StringValue.String, actualValue.StringValue.String)
			}
			if v.BoolValue.Status == pgtype.Present && (v.BoolValue.Bool != actualValue.BoolValue.Bool) {
				return fmt.Errorf("expected boolean value %v but got %v", v.BoolValue.Bool, actualValue.BoolValue.Bool)
			}
			if v.StringArrayValue.Status == pgtype.Present {
				expectedArr := database.FromTextArray(v.StringArrayValue)
				actualArr := database.FromTextArray(actualValue.StringArrayValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected string array %v but got %v", expectedArr, actualArr)
					}
				}
			}
			if v.IntArrayValue.Status == pgtype.Present {
				expectedArr := database.FromInt4Array(v.IntArrayValue)
				actualArr := database.FromInt4Array(actualValue.IntArrayValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected int array %v but got %v", expectedArr, actualArr)
					}
				}
			}
			if v.StringSetValue.Status == pgtype.Present {
				expectedArr := database.FromTextArray(v.StringSetValue)
				actualArr := database.FromTextArray(actualValue.StringSetValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected string set %v but got %v", expectedArr, actualArr)
					}
				}
			}
			if v.IntSetValue.Status == pgtype.Present {
				expectedArr := database.FromInt4Array(v.IntSetValue)
				actualArr := database.FromInt4Array(actualValue.IntSetValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected int set %v but got %v", expectedArr, actualArr)
					}
				}
			}
		}
	}
	return nil
}

func (s *Suite) LessonSchedulingStatusUpdatesTo(ctx context.Context, schedulingStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := (&lesson_repo.LessonRepo{})
	lesson, err := lessonRepo.GetLessonByID(ctx, s.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}

	if lesson.SchedulingStatus != lesson_domain.LessonSchedulingStatus(schedulingStatus) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", lesson_domain.LessonSchedulingStatus(schedulingStatus), lesson.SchedulingStatus)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updatesStatusInTheLessonIsValue(ctx context.Context, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	status := lesson_domain.LessonSchedulingStatus(value)

	if status != lesson_domain.LessonSchedulingStatusCanceled &&
		status != lesson_domain.LessonSchedulingStatusCompleted &&
		status != lesson_domain.LessonSchedulingStatusDraft &&
		status != lesson_domain.LessonSchedulingStatusPublished {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%s is not type of SchedulingStatus", value)
	}
	// update lesson
	lesson := &lesson_repo.Lesson{
		LessonID:         database.Text(stepState.CurrentLessonID),
		SchedulingStatus: database.Text(string(status)),
		UpdatedAt:        database.Timestamptz(time.Now()),
	}
	_, err := database.UpdateFields(ctx, lesson, s.BobDB.Exec, "lesson_id", []string{
		"scheduling_status",
		"updated_at",
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("got error when update lesson status: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) StudentHaveAttendanceInfo(ctx context.Context, atttendanceStatus, atttendanceNotice, atttendanceReason, atttendanceNote string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, ok := lpb.StudentAttendStatus_value[atttendanceStatus]
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid attendance status: %s", atttendanceStatus)
	}
	_, ok = lpb.StudentAttendanceNotice_value[atttendanceNotice]
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid attendance notice: %s", atttendanceNotice)
	}
	_, ok = lpb.StudentAttendanceReason_value[atttendanceReason]
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid attendance reason: %s", atttendanceReason)
	}

	query := fmt.Sprintf(`UPDATE lesson_members
		SET attendance_status = $2,
		attendance_notice = $3,
		attendance_reason = $4,
		attendance_note = $5,
		updated_at = now()
		WHERE lesson_id = $1`,
	)
	if _, err := s.BobDB.Exec(ctx, query, stepState.CurrentLessonID, atttendanceStatus, atttendanceNotice, atttendanceReason, atttendanceNote); err != nil {
		return ctx, fmt.Errorf("update member attendance: %v", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userHasBeenGrantedPermission(ctx context.Context, permissionName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	roleName := stepState.RoleName
	resourcePath, _ := interceptors.ResourcePathFromContext(ctx)

	query := `with 
	role as (
	  select r.role_id, r.resource_path
	  from "role" r
	  where r.role_name = $1
		and r.resource_path = $2
	),
	permission as (
		select p.permission_id, p.resource_path
	  	from "permission" p 
	  	where p.permission_name = $3
	)
	insert into permission_role
	  (permission_id, role_id, created_at, updated_at, resource_path)
	select permission.permission_id,role.role_id, now(), now(), role.resource_path
	  from role, permission
	  where role.resource_path = permission.resource_path
	  on conflict on constraint permission_role__pk do nothing; `

	_, err := s.BobDB.Exec(ctx, query, roleName, resourcePath, permissionName)
	if err != nil {
		return nil, fmt.Errorf("cannot grant %s permission for this user: %w", permissionName, err)
	}
	return StepStateToContext(ctx, stepState), nil
}
