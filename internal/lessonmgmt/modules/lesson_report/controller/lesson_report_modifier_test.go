package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lesson_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockLessonModifierService struct {
	mock.Mock
}

func (r *MockLessonModifierService) UpdateLessonSchedulingStatus(arg1 context.Context, arg2 *lpb.UpdateLessonSchedulingStatusRequest) (*lpb.UpdateLessonSchedulingStatusResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.UpdateLessonSchedulingStatusResponse), args.Error(1)
}
func TestLessonReportModifierService_SaveDraftGroupLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	unleashClientIns := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClientIns, "local")
	lessonRepo := new(mock_lesson_repositories.MockLessonRepo)
	lessonMemberRepo := new(mock_lesson_repositories.MockLessonMemberRepo)
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonModifierService := new(MockLessonModifierService)
	reallocationRepo := &mock_lesson_repositories.MockReallocationRepo{}
	masterDataRepo := &mock_lesson_repositories.MockMasterDataRepo{}

	partnerFormConfigs := `
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
						"is_required": true,
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
					}
				]
			}
		]
	}
`
	tcs := []struct {
		name     string
		req      *lpb.WriteGroupLessonReportRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "submit new draft lesson report",
			req: &lpb.WriteGroupLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*lpb.GroupLessonReportDetails{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.
					On("FindByLessonID", ctx, db, "lesson-id-1").
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameIndividualLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-id-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-id-1",
					StudentID:            "student-id-1",
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
				}}
				lessonReportRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))
				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteGroupLessonReportRequest{
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.GroupLessonReportDetails{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-1",
					StudentID:            "student-id-1",
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
				}}
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteGroupLessonReportRequest{
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.GroupLessonReportDetails{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			hasError: true,
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db

				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(nil, fmt.Errorf("err")).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)
			srv := NewLessonReportModifierService(
				wrapperConnection,
				lessonRepo,
				lessonMemberRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				partnerFormConfigRepo,
				reallocationRepo,
				lessonModifierService.UpdateLessonSchedulingStatus,
				unleashClientIns,
				"staging",
				masterDataRepo,
			)
			res, err := srv.SaveDraftGroupLessonReport(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.LessonReportId)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				lessonRepo,
				lessonMemberRepo,
				partnerFormConfigRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				unleashClientIns,
			)
		})
	}
}
func TestLessonReportModifierService_SubmitGroupLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	unleashClientIns := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClientIns, "local")
	lessonRepo := new(mock_lesson_repositories.MockLessonRepo)
	lessonMemberRepo := new(mock_lesson_repositories.MockLessonMemberRepo)
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonModifierService := new(MockLessonModifierService)
	reallocationRepo := &mock_lesson_repositories.MockReallocationRepo{}
	masterDataRepo := &mock_lesson_repositories.MockMasterDataRepo{}

	partnerFormConfigs := `
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
						"is_required": true,
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
					}
				]
			}
		]
	}
`
	tcs := []struct {
		name     string
		req      *lpb.WriteGroupLessonReportRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "submit new draft lesson report",
			req: &lpb.WriteGroupLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*lpb.GroupLessonReportDetails{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
						ReportVersion: 1,
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
						ReportVersion: 1,
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.
					On("FindByLessonID", ctx, db, "lesson-id-1").
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameIndividualLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-id-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-id-1",
					StudentID:            "student-id-1",
					ReportVersion:        1,
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
					ReportVersion:        1,
				}}
				lessonReportRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertWithVersion", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetReportVersionByLessonID", ctx, db, mock.Anything).
					Return(domain.LessonReportDetails{
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-1", StudentID: "student-id-1", ReportVersion: 1},
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-2", StudentID: "student-id-2", ReportVersion: 1},
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteGroupLessonReportRequest{
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.GroupLessonReportDetails{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
						ReportVersion: 1,
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
						ReportVersion: 1,
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-1",
					StudentID:            "student-id-1",
					ReportVersion:        1,
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
					ReportVersion:        1,
				}}

				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetReportVersionByLessonID", ctx, db, mock.Anything).
					Return(domain.LessonReportDetails{
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-1", StudentID: "student-id-1", ReportVersion: 1},
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-2", StudentID: "student-id-2", ReportVersion: 1},
					}, nil).
					Once()
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertWithVersion", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()

				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteGroupLessonReportRequest{
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.GroupLessonReportDetails{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
						ReportVersion: 1,
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
						ReportVersion: 1,
					},
				},
			},
			hasError: true,
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetReportVersionByLessonID", ctx, db, mock.Anything).
					Return(domain.LessonReportDetails{
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-1", StudentID: "student-id-1", ReportVersion: 1},
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-2", StudentID: "student-id-2", ReportVersion: 1},
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db

				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(nil, fmt.Errorf("err")).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)
			srv := NewLessonReportModifierService(
				wrapperConnection,
				lessonRepo,
				lessonMemberRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				partnerFormConfigRepo,
				reallocationRepo,
				lessonModifierService.UpdateLessonSchedulingStatus,
				unleashClientIns,
				"staging",
				masterDataRepo,
			)
			res, err := srv.SubmitGroupLessonReport(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.LessonReportId)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				lessonRepo,
				lessonMemberRepo,
				partnerFormConfigRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				unleashClientIns,
			)
		})
	}
}
func TestLessonReportModifierService_SubmitIndividualLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	unleashClientIns := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClientIns, "local")
	lessonRepo := new(mock_lesson_repositories.MockLessonRepo)
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	lessonMemberRepo := new(mock_lesson_repositories.MockLessonMemberRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonModifierService := new(MockLessonModifierService)
	reallocationRepo := &mock_lesson_repositories.MockReallocationRepo{}
	masterDataRepo := &mock_lesson_repositories.MockMasterDataRepo{}

	partnerFormConfigs := `
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
						"is_required": true,
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
					}
				]
			}
		]
	}
`
	tcs := []struct {
		name     string
		req      *lpb.WriteIndividualLessonReportRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "submit new individual lesson report",
			req: &lpb.WriteIndividualLessonReportRequest{
				FeatureName: string(lesson_report_consts.FeatureNameGroupLessonReport),
				LessonId:    "lesson-id-1",
				Details: []*lpb.IndividualLessonReportDetail{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				lessonReportRepo.
					On("FindByLessonID", ctx, db, "lesson-id-1").
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameIndividualLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-id-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-id-1",
					StudentID:            "student-id-1",
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
				}}
				lessonReportRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertWithVersion", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetReportVersionByLessonID", ctx, db, mock.Anything).
					Return(domain.LessonReportDetails{
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-1", StudentID: "student-id-1", ReportVersion: 0},
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-2", StudentID: "student-id-2", ReportVersion: 0},
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteIndividualLessonReportRequest{
				FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.IndividualLessonReportDetail{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
						ReportVersion: 1,
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
						ReportVersion: 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-1",
					StudentID:            "student-id-1",
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
				}}
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertWithVersion", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetReportVersionByLessonID", ctx, db, mock.Anything).
					Return(domain.LessonReportDetails{
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-1", StudentID: "student-id-1", ReportVersion: 1},
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-2", StudentID: "student-id-2", ReportVersion: 0},
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "submit lesson report",
			req: &lpb.WriteIndividualLessonReportRequest{
				FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.IndividualLessonReportDetail{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
						ReportVersion: 1,
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
						ReportVersion: 0,
					},
				},
			},
			hasError: true,
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetReportVersionByLessonID", ctx, db, mock.Anything).
					Return(domain.LessonReportDetails{
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-1", StudentID: "student-id-1", ReportVersion: 1},
						&domain.LessonReportDetail{LessonReportDetailID: "lesson-report-2", StudentID: "student-id-2", ReportVersion: 0},
					}, nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db

				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(nil, fmt.Errorf("err")).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)
			srv := NewLessonReportModifierService(
				wrapperConnection,
				lessonRepo,
				lessonMemberRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				partnerFormConfigRepo,
				reallocationRepo,
				lessonModifierService.UpdateLessonSchedulingStatus,
				unleashClientIns,
				"staging",
				masterDataRepo,
			)
			res, err := srv.SubmitIndividualLessonReport(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.LessonReportId)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				lessonRepo,
				partnerFormConfigRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				unleashClientIns,
			)
		})
	}
}
func TestLessonReportModifierService_SaveDraftIndividualLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	unleashClientIns := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClientIns, "local")
	lessonRepo := new(mock_lesson_repositories.MockLessonRepo)
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	lessonMemberRepo := new(mock_lesson_repositories.MockLessonMemberRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonModifierService := new(MockLessonModifierService)
	reallocationRepo := &mock_lesson_repositories.MockReallocationRepo{}
	masterDataRepo := &mock_lesson_repositories.MockMasterDataRepo{}

	partnerFormConfigs := `
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
						"is_required": true,
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
					}
				]
			}
		]
	}
`
	tcs := []struct {
		name     string
		req      *lpb.WriteIndividualLessonReportRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "submit new draft lesson report",
			req: &lpb.WriteIndividualLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*lpb.IndividualLessonReportDetail{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.
					On("FindByLessonID", ctx, db, "lesson-id-1").
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameIndividualLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-id-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-id-1",
					StudentID:            "student-id-1",
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
				}}
				lessonReportRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteIndividualLessonReportRequest{
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.IndividualLessonReportDetail{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db
				report := &domain.LessonReport{
					LessonReportID: "lesson-report-1",
				}
				details := domain.LessonReportDetails{{
					LessonReportDetailID: "detail-1",
					LessonReportID:       "lesson-report-1",
					StudentID:            "student-id-1",
				}, {
					LessonReportDetailID: "detail-2",
					LessonReportID:       "lesson-report-id-2",
					StudentID:            "student-id-2",
				}}
				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, mock.Anything).
					Return(domain.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				partnerFormConfigRepo.
					On("DeleteByLessonReportDetailIDs", ctx, tx, mock.Anything).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.Anything,
						repo.UpdateLessonMemberFields{
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceRemark),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFieldAttendanceStatus),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNotice),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceReason),
							repo.UpdateLessonMemberField(lesson_report_consts.SystemDefinedFiledAttendanceNote),
						}).Run(func(args mock.Arguments) {
					members := args[2].([]*lesson_domain.LessonMember)
					expected := []*lesson_domain.LessonMember{
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-1",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
						{
							LessonID:         "lesson-id-1",
							StudentID:        "student-id-2",
							AttendanceRemark: "",
							AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY.String(),
							AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE.String(),
							AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON.String(),
							AttendanceNote:   "lazy",
						},
					}
					assert.Equal(t, len(expected), len(members))

				}).Return(nil)
			},
		},
		{
			name: "save draft lesson report",
			req: &lpb.WriteIndividualLessonReportRequest{
				LessonId:       "lesson-id-1",
				LessonReportId: "lesson-report-1",
				Details: []*lpb.IndividualLessonReportDetail{
					{
						StudentId: "student-id-1",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &lpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &lpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &lpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &lpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &lpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &lpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &lpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &lpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &lpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId: "student-id-2",
						FieldValues: []*lpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &lpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &lpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			hasError: true,
			setup: func(ctx context.Context) {
				unleashClientIns.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClientIns.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				lessonReportRepo.On("FindByID", ctx, db, "lesson-report-1").
					Return(&domain.LessonReport{
						LessonReportID: "lesson-report-1",
						LessonID:       "lesson-id-1",
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(&lesson_domain.Lesson{
						LessonID:       "lesson-id-1",
						TeachingMethod: lesson_domain.LessonTeachingMethodGroup,
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, 1, string(lesson_report_consts.FeatureNameGroupLessonReport)).
					Return(&domain.PartnerFormConfig{
						FormConfigID:   "form-config-id-1",
						PartnerID:      1,
						FeatureName:    string(lesson_report_consts.FeatureNameGroupLessonReport),
						FormConfigData: []byte(partnerFormConfigs),
					}, nil).
					Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				// mock repo methods which will be called to store data to db

				lessonReportRepo.
					On("Update", ctx, tx, mock.Anything).
					Return(nil, fmt.Errorf("err")).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)
			srv := NewLessonReportModifierService(
				wrapperConnection,
				lessonRepo,
				lessonMemberRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				partnerFormConfigRepo,
				reallocationRepo,
				lessonModifierService.UpdateLessonSchedulingStatus,
				unleashClientIns,
				"staging",
				masterDataRepo,
			)
			res, err := srv.SaveDraftIndividualLessonReport(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, res.LessonReportId)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				lessonRepo,
				partnerFormConfigRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				unleashClientIns,
			)
		})
	}
}
