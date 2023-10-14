package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLessonReportModifierService_SubmitLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonModifierService := new(MockLessonModifierService)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	tcs := []struct {
		name     string
		req      *bpb.WriteLessonReportRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "submit new lesson report",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:         database.Text("lesson-id-1"),
						TeacherID:        database.Text("teacher-id-1"),
						SchedulingStatus: database.Text(string(entities.LessonSchedulingStatusPublished)),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Create", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.NotEmpty(t, e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSubmitted, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSubmitted))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(15),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				lessonModifierService.On("UpdateLessonSchedulingStatus", ctx, mock.MatchedBy(func(req *lpb.UpdateLessonSchedulingStatusRequest) bool {
					assert.Equal(t, "lesson-id-1", req.LessonId)
					assert.Equal(t, cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED, req.SchedulingStatus)
					return true
				})).Return(&lpb.UpdateLessonSchedulingStatusResponse{}, nil).Once()
			},
		},
		{
			name: "submit new lesson report with a canceled lesson",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:         database.Text("lesson-id-1"),
						TeacherID:        database.Text("teacher-id-1"),
						SchedulingStatus: database.Text(string(entities.LessonSchedulingStatusCanceled)),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Create", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.NotEmpty(t, e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSubmitted, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSubmitted))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(15),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
			},
		},
		{
			name: "submit to update a lesson report",
			req: &bpb.WriteLessonReportRequest{
				LessonReportId: "lesson-report-id-1",
				LessonId:       "lesson-id-2", // lesson id is wrong, so it will be replaced by another id
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByID", ctx, db, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSaved)), // report submitting status will be replaced
						FormConfigID:           database.Text("form-config-id-2"),                           // form config id will be replaced
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Update", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.Equal(t, "lesson-report-id-1", e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSubmitted, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSubmitted))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(15),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
			},
		},
		{
			name: "submit new lesson report with missing required fields not update status with unleash false",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
			},
			hasError: true,
		},
		{
			name: "submit new lesson report with non-existed field",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "non-existed-id",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
			},
			hasError: true,
		},
		{
			name: "submit new lesson report when status lesson is publish",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:         database.Text("lesson-id-1"),
						TeacherID:        database.Text("teacher-id-1"),
						SchedulingStatus: database.Text(string(entities.LessonSchedulingStatusPublished)),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Create", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.NotEmpty(t, e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSubmitted, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSubmitted))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(15),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
		},
		{
			name: "submit new lesson report with student attendance not registered",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "non-existed-id",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := &LessonReportModifierService{
				DB:                           db,
				PartnerFormConfigRepo:        partnerFormConfigRepo,
				LessonRepo:                   lessonRepo,
				LessonReportRepo:             lessonReportRepo,
				LessonReportDetailRepo:       lessonReportDetailRepo,
				LessonMemberRepo:             lessonMemberRepo,
				TeacherRepo:                  teacherRepo,
				UpdateLessonSchedulingStatus: lessonModifierService.UpdateLessonSchedulingStatus,
				UnleashClientIns:             mockUnleashClient,
				Env:                          "local",
			}
			res, err := srv.SubmitLessonReport(ctx, tc.req)
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
				teacherRepo,
				partnerFormConfigRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				lessonMemberRepo,
				lessonModifierService,
			)
		})
	}
}

func TestLessonReportModifierService_SaveDraftLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)

	tcs := []struct {
		name     string
		req      *bpb.WriteLessonReportRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "submit new draft lesson report",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Create", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.NotEmpty(t, e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSaved, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSaved))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(15),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
								CreatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								UpdatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								DeleteAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
								CourseID: pgtype.Text{
									Status: pgtype.Null,
								},
								CreatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								UpdatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								DeleteAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
			},
		},
		{
			name: "save a draft lesson report",
			req: &bpb.WriteLessonReportRequest{
				LessonReportId: "lesson-report-id-1",
				LessonId:       "lesson-id-2", // lesson id is wrong, so it will be replaced by another id
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByID", ctx, db, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSubmitted)), // report submitting status will be replaced
						FormConfigID:           database.Text("form-config-id-2"),                               // form config id will be replaced
					}, nil).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Update", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.Equal(t, "lesson-report-id-1", e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSaved, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSaved))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(15),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
								CreatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								UpdatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								DeleteAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
								CourseID: pgtype.Text{
									Status: pgtype.Null,
								},
								CreatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								UpdatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								DeleteAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
			},
		},
		{
			name: "submit new lesson report with missing required fields",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()

				// mock repo methods which will be called to store data to db
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("Create", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						e := args[2].(*entities.LessonReport)
						assert.NotEmpty(t, e.LessonReportID.String)
						assert.Equal(t, "lesson-id-1", e.LessonID.String)
						assert.EqualValues(t, entities.ReportSubmittingStatusSaved, e.ReportSubmittingStatus.String)
						assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

						report.LessonReportID = e.LessonReportID
						report.LessonID = database.Text("lesson-id-1")
						report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSaved))
						report.FormConfigID = database.Text("form-config-id-1")
					}).
					Return(report, nil).
					Once()
				lessonReportDetailRepo.
					On("Upsert", ctx, db, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text)
						assert.NotEmpty(t, id.String)
						assert.Equal(t, report.LessonReportID.String, id.String)
						actualDetails := args[3].(entities.LessonReportDetails)
						assert.Len(t, actualDetails, 2)
						expected := entities.LessonReportDetails{
							{
								StudentID: database.Text("student-id-1"),
							},
							{
								StudentID: database.Text("student-id-2"),
							},
						}
						for i, detail := range actualDetails {
							assert.NotEmpty(t, detail.LessonReportID.String)
							assert.NotEmpty(t, detail.LessonReportDetailID.String)
							assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
							assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

							details[i].LessonReportDetailID = detail.LessonReportDetailID
							details[i].LessonReportID = detail.LessonReportID
							details[i].StudentID = detail.StudentID
						}
					}).
					Return(nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args[2].(pgtype.Text).String
						assert.NotEmpty(t, id)
						assert.Equal(t, report.LessonReportID.String, id)
					}).
					Return(entities.LessonReportDetails{
						details[0],
						details[1],
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("UpsertFieldValues", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
						expected := []*entities.PartnerDynamicFormFieldValue{
							{
								FieldID:          database.Text("ordinal-number"),
								IntValue:         database.Int4(5),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("title"),
								StringValue:      database.Text("monitor"),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("scores"),
								IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
							{
								FieldID:          database.Text("comments"),
								StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
							},
							{
								FieldID:        database.Text("buddy"),
								StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
							},
							{
								FieldID:     database.Text("finished-exams"),
								IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
							},
							{
								FieldID:          database.Text("is-pass-lesson"),
								BoolValue:        database.Bool(true),
								FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
							},
						}

						assert.Len(t, fieldVals, len(expected))
						for i, field := range fieldVals {
							assert.NotEmpty(t, field.LessonReportDetailID.String)
							assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
							assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
							if expected[i].IntValue.Status != pgtype.Present {
								assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
							}
							if expected[i].StringValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
							}
							if expected[i].BoolValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
							}
							if expected[i].IntArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
							}
							if expected[i].StringArrayValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
							}
							if expected[i].StringSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
							}
							if expected[i].IntSetValue.Status == pgtype.Present {
								assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
							}
							assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
						}
					}).
					Return(nil).
					Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, db, mock.Anything, entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
						entities.LessonMemberAttendanceNotice,
						entities.LessonMemberAttendanceReason,
						entities.LessonMemberAttendanceNote,
					}).
					Run(func(args mock.Arguments) {
						members := args[2].([]*entities.LessonMember)
						expected := []*entities.LessonMember{
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-1"),
								AttendanceRemark: database.Text("very good"),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
								CreatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								UpdatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								DeleteAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
							},
							{
								LessonID:         database.Text("lesson-id-1"),
								UserID:           database.Text("student-id-2"),
								AttendanceRemark: database.Text(""),
								AttendanceStatus: database.Text(string(entities.StudentAttendStatusLeaveEarly)),
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeOnTheDay)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonSchoolEvent)),
								AttendanceNote:   database.Text("lazy"),
								CourseID: pgtype.Text{
									Status: pgtype.Null,
								},
								CreatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								UpdatedAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
								DeleteAt: pgtype.Timestamptz{
									Status: pgtype.Null,
								},
							},
						}
						assert.Equal(t, len(expected), len(members))
						for i, member := range expected {
							assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
							assert.Equal(t, member.UserID.String, members[i].UserID.String)
							assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
							assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
							assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
							assert.NotZero(t, members[i].CreatedAt.Time)
							assert.NotZero(t, members[i].UpdatedAt.Time)
						}
					}).
					Return(nil).Once()
			},
			hasError: false,
		},
		{
			name: "submit new lesson report with non-existed field",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id-1",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "non-existed-id",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_ON_THE_DAY,
						AttendanceReason: bpb.StudentAttendanceReason_SCHOOL_EVENT,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
				// mock repo methods which will be called to normalize data
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
				teacherRepo.
					On("FindByID", ctx, db, database.Text("teacher-id-1")).
					Return(&entities.Teacher{
						ID:        database.Text("teacher-id-1"),
						SchoolIDs: database.Int4Array([]int32{1, 3}),
					}, nil).
					Once()
				partnerFormConfigRepo.
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameIndividualLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()

				// mock repo methods which will be called to validate data
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := &LessonReportModifierService{
				DB:                     db,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonRepo:             lessonRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			}
			res, err := srv.SaveDraftLessonReport(ctx, tc.req)
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
				teacherRepo,
				partnerFormConfigRepo,
				lessonReportRepo,
				lessonReportDetailRepo,
				lessonMemberRepo,
			)
		})
	}
}

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
