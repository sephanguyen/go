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
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewLessonReport_ByLessonReportGRPCMessage(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		req      *bpb.WriteLessonReportRequest
		expected *LessonReport
	}{
		{
			name: "new lesson report with full fields without feature name",
			req: &bpb.WriteLessonReportRequest{
				LessonReportId: "lesson-report-id",
				LessonId:       "lesson-id",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: bpb.StudentAttendanceReason_FAMILY_REASON,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_STRING_SET,
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT_SET,
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: bpb.StudentAttendanceReason_FAMILY_REASON,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{7, 5},
									},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Gabriel"},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_STRING_SET,
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT_SET,
							},
						},
					},
				},
			},
			expected: &LessonReport{
				LessonReportID: "lesson-report-id",
				LessonID:       "lesson-id",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_STRING.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_BOOL.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String(),
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_STRING_SET.String(),
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT_SET.String(),
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_BOOL.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{7, 5},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Gabriel"},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_STRING_SET.String(),
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT_SET.String(),
							},
						},
					},
				},
			},
		},
		{
			name: "new lesson report with full fields",
			req: &bpb.WriteLessonReportRequest{
				FeatureName:    "test-feature-name",
				LessonReportId: "lesson-report-id",
				LessonId:       "lesson-id",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: bpb.StudentAttendanceReason_FAMILY_REASON,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(5),
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "title",
								Value: &bpb.DynamicFieldValue_StringValue{
									StringValue: "monitor",
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{9, 10, 8, 10},
									},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "comments",
								Value: &bpb.DynamicFieldValue_StringArrayValue_{
									StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
										ArrayValue: []string{"excellent", "creative", "diligent"},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT,
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT,
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2, 3, 5, 6, 1},
									},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT,
							},
						},
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: bpb.StudentAttendanceReason_FAMILY_REASON,
						AttendanceNote:   "lazy",
						FieldValues: []*bpb.DynamicFieldValue{
							{
								DynamicFieldId: "ordinal-number",
								Value: &bpb.DynamicFieldValue_IntValue{
									IntValue: int32(15),
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "is-pass-lesson",
								Value: &bpb.DynamicFieldValue_BoolValue{
									BoolValue: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{7, 5},
									},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT,
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "buddy",
								ValueType:      bpb.ValueType_VALUE_TYPE_INT,
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								ValueType:      bpb.ValueType_VALUE_TYPE_INT,
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2},
									},
								},
							},
						},
					},
				},
			},
			expected: &LessonReport{
				LessonReportID: "lesson-report-id",
				LessonID:       "lesson-id",
				FeatureName:    "test-feature-name",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID:   "comments",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID:   "buddy",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID:   "finished-exams",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{7, 5},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Gabriel"},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2},
								},
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
							},
						},
					},
				},
			},
		},
		{
			name: "new lesson report with missing lesson_report_id, course_id, attendances field",
			req: &bpb.WriteLessonReportRequest{
				LessonId: "lesson-id",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId: "student-id-1",
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
						StudentId: "student-id-2",
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
							{
								DynamicFieldId: "scores",
								Value: &bpb.DynamicFieldValue_IntArrayValue_{
									IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
										ArrayValue: []int32{7, 5},
									},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								DynamicFieldId: "buddy",
								Value: &bpb.DynamicFieldValue_StringSetValue_{
									StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
										ArrayValue: []string{"Gabriel"},
									},
								},
							},
							{
								DynamicFieldId: "finished-exams",
								Value: &bpb.DynamicFieldValue_IntSetValue_{
									IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
										ArrayValue: []int32{1, 2},
									},
								},
							},
						},
					},
				},
			},
			expected: &LessonReport{
				LessonID: "lesson-id",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusEmpty,
						AttendanceNotice: entities.StudentAttendanceNoticeEmpty,
						AttendanceReason: entities.StudentAttendanceReasonEmpty,
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID:   "comments",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID:   "buddy",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID:   "finished-exams",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusEmpty,
						AttendanceNotice: entities.StudentAttendanceNoticeEmpty,
						AttendanceReason: entities.StudentAttendanceReasonEmpty,
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{7, 5},
								},
								ValueType:        bpb.ValueType_VALUE_TYPE_INT.String(),
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID:   "buddy",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									StringSet: []string{"Gabriel"},
								},
							},
							{
								FieldID:   "finished-exams",
								ValueType: bpb.ValueType_VALUE_TYPE_INT.String(),
								Value: &AttributeValue{
									IntSet: []int{1, 2},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "new lesson report without details",
			req: &bpb.WriteLessonReportRequest{
				LessonReportId: "lesson-report-id",
				LessonId:       "lesson-id",
			},
			expected: &LessonReport{
				LessonReportID: "lesson-report-id",
				LessonID:       "lesson-id",
			},
		},
		{
			name: "new lesson report without any fields in details",
			req: &bpb.WriteLessonReportRequest{
				LessonReportId: "lesson-report-id",
				LessonId:       "lesson-id",
				Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
					{
						StudentId:        "student-id-1",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
						AttendanceRemark: "very good",
						AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: bpb.StudentAttendanceReason_FAMILY_REASON,
						AttendanceNote:   "lazy",
					},
					{
						StudentId:        "student-id-2",
						AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
						AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
						AttendanceReason: bpb.StudentAttendanceReason_FAMILY_REASON,
						AttendanceNote:   "lazy",
					},
				},
			},
			expected: &LessonReport{
				LessonReportID: "lesson-report-id",
				LessonID:       "lesson-id",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields:           LessonReportFields{},
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						Fields:           LessonReportFields{},
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLessonReport(ByLessonReportGRPCMessage(tc.req))
			require.NoError(t, err)
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}

func TestLessonReport_Normalize(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	lessonRepo := new(mock_repositories.MockLessonRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	partnerFormConfigRepo := new(mock_repositories.MockPartnerFormConfigRepo)

	tcs := []struct {
		name         string
		lessonReport *LessonReport
		expected     *LessonReport
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name: "normalize lesson report with lesson have teaching method is group",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusApproved,
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:       database.Text("lesson-id-1"),
						TeacherID:      database.Text("teacher-id-1"),
						TeachingMethod: database.Text(string(entities.LessonTeachingMethodGroup)),
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
					// On("FindByFeatureName", ctx, db, database.Text(string(entities.FeatureNameGroupLessonReport))).
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameGroupLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID: database.Text("form-config-id-1"),
						PartnerID:    database.Int4(1),
						FeatureName:  database.Text(string(entities.FeatureNameGroupLessonReport)),
						FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
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
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
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
									"field_id": "field-id-3",
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
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
					}, nil).
					Once()
			},
			expected: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusApproved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "field-id-1",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "field-id-2",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "field-id-3",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "field-id-4",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "normalize lesson report when missing lesson report id",
			lessonReport: &LessonReport{
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusApproved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
			},
			expected: &LessonReport{
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusApproved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
			},
		},
		{
			name: "normalize lesson report when missing submitting status",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-1",
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
			},
			expected: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSaved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
			},
		},
		{
			name: "normalize lesson report when form config is empty",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusApproved,
			},
			setup: func(ctx context.Context) {
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
						FormConfigID: database.Text("form-config-id-1"),
						PartnerID:    database.Int4(1),
						FeatureName:  database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
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
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
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
									"field_id": "field-id-3",
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
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
					}, nil).
					Once()
			},
			expected: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusApproved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "field-id-1",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "field-id-2",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "field-id-3",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "field-id-4",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "normalize lesson report when have some duplicated details",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-1",
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
				Details: LessonReportDetails{
					{
						StudentID: "student-id-1",
					},
					{
						StudentID: "student-id-2",
					},
					{
						StudentID: "student-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
			},
			expected: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSaved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
				Details: LessonReportDetails{
					{
						StudentID: "student-id-1",
					},
					{
						StudentID: "student-id-2",
					},
				},
			},
		},
		{
			name: "normalize lesson report when some details have duplicated fields",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-1",
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "is-not-monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: false,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{7, 5},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Gabriel"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, nil).
					Once()
			},
			expected: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSaved,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{7, 5},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Gabriel"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "normalize lesson report with non-exist lesson id",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-1",
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id-1",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.Lesson{
						LessonID:  database.Text("lesson-id-1"),
						TeacherID: database.Text("teacher-id-1"),
					}, fmt.Errorf("could not find lesson")).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			tc.lessonReport.LessonRepo = lessonRepo
			tc.lessonReport.PartnerFormConfigRepo = partnerFormConfigRepo
			tc.lessonReport.TeacherRepo = teacherRepo

			err := tc.lessonReport.Normalize(ctx, db)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.LessonReportID, tc.lessonReport.LessonReportID)
				assert.Equal(t, tc.expected.LessonID, tc.lessonReport.LessonID)
				assert.EqualValues(t, tc.expected.SubmittingStatus, tc.lessonReport.SubmittingStatus)
				assert.EqualValues(t, tc.expected.FormConfig, tc.lessonReport.FormConfig)
				assert.EqualValues(t, tc.expected.Details, tc.lessonReport.Details)
				mock.AssertExpectationsForObjects(t, db, lessonRepo, teacherRepo, partnerFormConfigRepo)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, teacherRepo, partnerFormConfigRepo)
		})
	}
}

func TestLessonReport_IsValid(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	lessonRepo := new(mock_repositories.MockLessonRepo)

	tcs := []struct {
		name         string
		lessonReport *LessonReport
		setup        func(ctx context.Context)
		isValid      bool
		isValidDraft bool
	}{
		{
			name: "validate lesson report with full fields",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceNotice),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceReason),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceNote),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report without details",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report without form config",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report without detail's fields, attendance_status, attendance_remark and form config data",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
				},
				Details: LessonReportDetails{
					{
						StudentID: "student-id-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report without detail's fields and form config data",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report without detail's fields, attendance_status, attendance_remark and form config",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				Details: LessonReportDetails{
					{
						StudentID: "student-id-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report without detail's fields and form config",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report without lesson id",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
				},
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report without submitting status",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-1",
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
				},
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report with detail's student id is empty",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
					},
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report with detail's student is duplicated",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
					},
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report with detail's fields id is empty",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report with detail's fields id is duplicated",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:   "ordinal-number",
										ValueType: "VALUE_TYPE_INT",
									},
									{
										FieldID:   "title",
										ValueType: "VALUE_TYPE_STRING",
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:   "comments",
										ValueType: "VALUE_TYPE_STRING_ARRAY",
									},
									{
										FieldID:   "finished-exams",
										ValueType: "VALUE_TYPE_INT_ARRAY",
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report with non-existed field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:   "ordinal-number",
										ValueType: "VALUE_TYPE_INT",
									},
									{
										FieldID:   "title",
										ValueType: "VALUE_TYPE_STRING",
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "non-existed-id",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report with missing not required field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report with missing required field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: true,
		},
		{
			name: "validate lesson report with missing attendance_status and attendance_remark which is not required field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID: "student-id-2",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report with missing attendance_status which is required field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusEmpty,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusEmpty,
						AttendanceRemark: "good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: true,
		},
		{
			name: "validate lesson report with missing attendance_remark which is required field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusLate,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: true,
		},
		{
			name: "validate lesson report with attendance_status and attendance_remark which is required field",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusLate,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceRemark: "good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      true,
			isValidDraft: true,
		},
		{
			name: "validate lesson report with some fields have id same system defined fields",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: string(SystemDefinedFieldAttendanceStatus),
								Value: &AttributeValue{
									String: "status",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
		{
			name: "validate lesson report when have some details contain student id not belong to lesson",
			lessonReport: &LessonReport{
				LessonReportID:   "lesson-report-id-1",
				LessonID:         "lesson-id-1",
				SubmittingStatus: entities.ReportSubmittingStatusSubmitted,
				FormConfig: &FormConfig{
					FormConfigID: "form-config-id",
					FormConfigData: &FormConfigData{
						Sections: []*FormConfigSection{
							{
								SectionID: "section-id-0",
								Fields: []*FormConfigField{
									{
										FieldID:    string(SystemDefinedFieldAttendanceStatus),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
									{
										FieldID:    string(SystemDefinedFieldAttendanceRemark),
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: false,
									},
								},
							},
							{
								SectionID: "section-id-1",
								Fields: []*FormConfigField{
									{
										FieldID:    "ordinal-number",
										ValueType:  "VALUE_TYPE_INT",
										IsRequired: true,
									},
									{
										FieldID:    "title",
										ValueType:  "VALUE_TYPE_STRING",
										IsRequired: true,
									},
								},
							},
							{
								SectionID: "section-id-2",
								Fields: []*FormConfigField{
									{
										FieldID:    "comments",
										ValueType:  "VALUE_TYPE_STRING_SET",
										IsRequired: false,
									},
									{
										FieldID:    "finished-exams",
										ValueType:  "VALUE_TYPE_INT_ARRAY",
										IsRequired: false,
									},
								},
							},
						},
					},
				},
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-5",
						AttendanceStatus: entities.StudentAttendStatusEmpty,
						AttendanceRemark: "very good",
					},
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Twice()
			},
			isValid:      false,
			isValidDraft: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(ctx)
			}
			tc.lessonReport.LessonRepo = lessonRepo
			err := tc.lessonReport.IsValid(ctx, db, false)
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			err = tc.lessonReport.IsValid(ctx, db, true)
			if tc.isValidDraft {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo)
		})
	}
}

const formConfigJson = `
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
				},
				{
					"field_id": "attendance_notice",
					"label": "attendance notice",
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
					"field_id": "attendance_reason",
					"label": "attendance reason",
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
					"field_id": "attendance_note",
					"label": "attendance note",
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
					"value_type": "VALUE_TYPE_INT_ARRAY",
					"is_required": false
				}
			]
		}
	]
}
`

func TestLessonReport_Submit(t *testing.T) {
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
		name         string
		lessonReport *LessonReport
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name: "submit new lesson report",
			lessonReport: &LessonReport{
				LessonID: "lesson-id-1",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
				LessonRepo:             lessonRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			},
			setup: func(ctx context.Context) {
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
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
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
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
					}).Run(func(args mock.Arguments) {
					members := args[2].([]*entities.LessonMember)
					expected := []*entities.LessonMember{
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-1"),
							AttendanceRemark: database.Text("very good"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
							AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
							AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
							AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
							AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
					}
					assert.Equal(t, len(expected), len(members))
					for i, member := range expected {
						assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
						assert.Equal(t, member.UserID.String, members[i].UserID.String)
						assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
						assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
						assert.Equal(t, member.AttendanceNotice.String, members[i].AttendanceNotice.String)
						assert.Equal(t, member.AttendanceReason.String, members[i].AttendanceReason.String)
						assert.Equal(t, member.AttendanceNote.String, members[i].AttendanceNote.String)
						assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
						assert.NotZero(t, members[i].CreatedAt.Time)
						assert.NotZero(t, members[i].UpdatedAt.Time)
					}
				}).Return(nil).Once()
			},
		},
		{
			name: "submit to update a lesson report",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-2", // lesson id is wrong, so it will be replaced by another id
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
				LessonRepo:             lessonRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			},
			setup: func(ctx context.Context) {
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("FindByID", ctx, db, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSaved)), // report submitting status will be replaced
						FormConfigID:           database.Text("form-config-id-2"),                           // form config id will be replaced
					}, nil).Once()
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
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
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
					}).Run(func(args mock.Arguments) {
					members := args[2].([]*entities.LessonMember)
					expected := []*entities.LessonMember{
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-1"),
							AttendanceRemark: database.Text("very good"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
							AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
							AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
							AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
							AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
					}
					assert.Equal(t, len(expected), len(members))
					for i, member := range expected {
						assert.Equal(t, member.LessonID.String, members[i].LessonID.String)
						assert.Equal(t, member.UserID.String, members[i].UserID.String)
						assert.Equal(t, member.AttendanceRemark.String, members[i].AttendanceRemark.String)
						assert.Equal(t, member.AttendanceStatus.String, members[i].AttendanceStatus.String)
						assert.Equal(t, member.AttendanceNotice.String, members[i].AttendanceNotice.String)
						assert.Equal(t, member.AttendanceReason.String, members[i].AttendanceReason.String)
						assert.Equal(t, member.AttendanceNote.String, members[i].AttendanceNote.String)
						assert.Equal(t, member.CourseID.String, members[i].CourseID.String)
						assert.NotZero(t, members[i].CreatedAt.Time)
						assert.NotZero(t, members[i].UpdatedAt.Time)
					}
				}).Return(nil).Once()
			},
		},
		// {
		// 	name: "submit new lesson report with locked lesson",
		// 	lessonReport: &LessonReport{
		// 		LessonID: "lesson-id-1",
		// 		Details: LessonReportDetails{
		// 			{
		// 				StudentID:        "student-id-1",
		// 				AttendanceStatus: entities.StudentAttendStatusAttend,
		// 				AttendanceRemark: "very good",
		// 				Fields: LessonReportFields{
		// 					{
		// 						FieldID: "ordinal-number",
		// 						Value: &AttributeValue{
		// 							Int: 5,
		// 						},
		// 						FieldRenderGuide: []byte("fake guide to render this field"),
		// 					},
		// 					{
		// 						FieldID: "title",
		// 						Value: &AttributeValue{
		// 							String: "monitor",
		// 						},
		// 						FieldRenderGuide: []byte("fake guide to render this field"),
		// 					},
		// 					{
		// 						FieldID: "is-pass-lesson",
		// 						Value: &AttributeValue{
		// 							Bool: true,
		// 						},
		// 						FieldRenderGuide: []byte("fake guide to render this field"),
		// 					},
		// 					{
		// 						FieldID: "scores",
		// 						Value: &AttributeValue{
		// 							IntArray: []int{9, 10, 8, 10},
		// 						},
		// 						FieldRenderGuide: []byte("fake guide to render this field"),
		// 					},
		// 					{
		// 						FieldID: "comments",
		// 						Value: &AttributeValue{
		// 							StringArray: []string{"excellent", "creative", "diligent"},
		// 						},
		// 					},
		// 					{
		// 						FieldID: "buddy",
		// 						Value: &AttributeValue{
		// 							StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
		// 						},
		// 					},
		// 					{
		// 						FieldID: "finished-exams",
		// 						Value: &AttributeValue{
		// 							IntSet: []int{1, 2, 3, 5, 6},
		// 						},
		// 					},
		// 				},
		// 			},
		// 			{
		// 				StudentID:        "student-id-2",
		// 				AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
		// 				Fields: LessonReportFields{
		// 					{
		// 						FieldID: "ordinal-number",
		// 						Value: &AttributeValue{
		// 							Int: 15,
		// 						},
		// 						FieldRenderGuide: []byte("fake guide to render this field"),
		// 					},
		// 					{
		// 						FieldID: "is-pass-lesson",
		// 						Value: &AttributeValue{
		// 							Bool: true,
		// 						},
		// 						FieldRenderGuide: []byte("fake guide to render this field"),
		// 					},
		// 				},
		// 			},
		// 		},
		// 		LessonRepo:             lessonRepo,
		// 		PartnerFormConfigRepo:  partnerFormConfigRepo,
		// 		LessonReportRepo:       lessonReportRepo,
		// 		LessonReportDetailRepo: lessonReportDetailRepo,
		// 		LessonMemberRepo:       lessonMemberRepo,
		// 		TeacherRepo:            teacherRepo,
		// 	},
		// 	setup: func(ctx context.Context) {
		// 		report := &entities.LessonReport{}
		// 		details := entities.LessonReportDetails{{}, {}}
		// 		lessonReportRepo.
		// 			On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
		// 			Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
		// 		lessonRepo.
		// 			On("FindByID", ctx, db, database.Text("lesson-id-1")).
		// 			Return(&entities.Lesson{
		// 				LessonID:  database.Text("lesson-id-1"),
		// 				TeacherID: database.Text("teacher-id-1"),
		// 				IsLocked:  database.Bool(true),
		// 			}, nil).
		// 			Once()
		// 		teacherRepo.
		// 			On("FindByID", ctx, db, database.Text("teacher-id-1")).
		// 			Return(&entities.Teacher{
		// 				ID:        database.Text("teacher-id-1"),
		// 				SchoolIDs: database.Int4Array([]int32{1, 3}),
		// 			}, nil).
		// 			Once()
		// 		partnerFormConfigRepo.
		// 			On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
		// 			Return(&entities.PartnerFormConfig{
		// 				FormConfigID:   database.Text("form-config-id-1"),
		// 				PartnerID:      database.Int4(1),
		// 				FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
		// 				FormConfigData: database.JSONB(formConfigJson),
		// 			}, nil).
		// 			Once()
		// 		lessonRepo.
		// 			On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
		// 			Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
		// 			Once()
		// 		lessonReportRepo.
		// 			On("Create", ctx, db, mock.Anything).
		// 			Run(func(args mock.Arguments) {
		// 				e := args[2].(*entities.LessonReport)
		// 				assert.NotEmpty(t, e.LessonReportID.String)
		// 				assert.Equal(t, "lesson-id-1", e.LessonID.String)
		// 				assert.EqualValues(t, entities.ReportSubmittingStatusSubmitted, e.ReportSubmittingStatus.String)
		// 				assert.EqualValues(t, "form-config-id-1", e.FormConfigID.String)

		// 				report.LessonReportID = e.LessonReportID
		// 				report.LessonID = database.Text("lesson-id-1")
		// 				report.ReportSubmittingStatus = database.Text(string(entities.ReportSubmittingStatusSubmitted))
		// 				report.FormConfigID = database.Text("form-config-id-1")
		// 			}).
		// 			Return(report, nil).
		// 			Once()
		// 		lessonReportDetailRepo.
		// 			On("Upsert", ctx, db, mock.Anything, mock.Anything).
		// 			Run(func(args mock.Arguments) {
		// 				id := args[2].(pgtype.Text)
		// 				assert.NotEmpty(t, id.String)
		// 				assert.Equal(t, report.LessonReportID.String, id.String)
		// 				actualDetails := args[3].(entities.LessonReportDetails)
		// 				assert.Len(t, actualDetails, 2)
		// 				expected := entities.LessonReportDetails{
		// 					{
		// 						StudentID: database.Text("student-id-1"),
		// 					},
		// 					{
		// 						StudentID: database.Text("student-id-2"),
		// 					},
		// 				}
		// 				for i, detail := range actualDetails {
		// 					assert.NotEmpty(t, detail.LessonReportID.String)
		// 					assert.NotEmpty(t, detail.LessonReportDetailID.String)
		// 					assert.Equal(t, report.LessonReportID.String, detail.LessonReportID.String)
		// 					assert.Equal(t, expected[i].StudentID.String, detail.StudentID.String)

		// 					details[i].LessonReportDetailID = detail.LessonReportDetailID
		// 					details[i].LessonReportID = detail.LessonReportID
		// 					details[i].StudentID = detail.StudentID
		// 				}
		// 			}).
		// 			Return(nil).
		// 			Once()
		// 		lessonReportDetailRepo.
		// 			On("GetByLessonReportID", ctx, db, mock.Anything).
		// 			Run(func(args mock.Arguments) {
		// 				id := args[2].(pgtype.Text).String
		// 				assert.NotEmpty(t, id)
		// 				assert.Equal(t, report.LessonReportID.String, id)
		// 			}).
		// 			Return(entities.LessonReportDetails{
		// 				details[0],
		// 				details[1],
		// 			}, nil).
		// 			Once()
		// 		lessonReportDetailRepo.
		// 			On("UpsertFieldValues", ctx, db, mock.Anything).
		// 			Run(func(args mock.Arguments) {
		// 				fieldVals := args[2].([]*entities.PartnerDynamicFormFieldValue)
		// 				expected := []*entities.PartnerDynamicFormFieldValue{
		// 					{
		// 						FieldID:          database.Text("ordinal-number"),
		// 						IntValue:         database.Int4(5),
		// 						FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
		// 					},
		// 					{
		// 						FieldID:          database.Text("title"),
		// 						StringValue:      database.Text("monitor"),
		// 						FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
		// 					},
		// 					{
		// 						FieldID:          database.Text("is-pass-lesson"),
		// 						BoolValue:        database.Bool(true),
		// 						FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
		// 					},
		// 					{
		// 						FieldID:          database.Text("scores"),
		// 						IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
		// 						FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
		// 					},
		// 					{
		// 						FieldID:          database.Text("comments"),
		// 						StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
		// 					},
		// 					{
		// 						FieldID:        database.Text("buddy"),
		// 						StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
		// 					},
		// 					{
		// 						FieldID:     database.Text("finished-exams"),
		// 						IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
		// 					},
		// 					{
		// 						FieldID:          database.Text("ordinal-number"),
		// 						IntValue:         database.Int4(15),
		// 						FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
		// 					},
		// 					{
		// 						FieldID:          database.Text("is-pass-lesson"),
		// 						BoolValue:        database.Bool(true),
		// 						FieldRenderGuide: database.JSONB([]byte("fake guide to render this field")),
		// 					},
		// 				}

		// 				assert.Len(t, fieldVals, len(expected))
		// 				for i, field := range fieldVals {
		// 					assert.NotEmpty(t, field.LessonReportDetailID.String)
		// 					assert.NotEmpty(t, field.DynamicFormFieldValueID.String)
		// 					assert.Equal(t, expected[i].FieldID.String, field.FieldID.String)
		// 					if expected[i].IntValue.Status != pgtype.Present {
		// 						assert.Equal(t, expected[i].IntValue.Int, field.IntValue.Int)
		// 					}
		// 					if expected[i].StringValue.Status == pgtype.Present {
		// 						assert.Equal(t, expected[i].StringValue.String, field.StringValue.String)
		// 					}
		// 					if expected[i].BoolValue.Status == pgtype.Present {
		// 						assert.Equal(t, expected[i].BoolValue.Bool, field.BoolValue.Bool)
		// 					}
		// 					if expected[i].IntArrayValue.Status == pgtype.Present {
		// 						assert.Equal(t, expected[i].IntArrayValue, field.IntArrayValue)
		// 					}
		// 					if expected[i].StringArrayValue.Status == pgtype.Present {
		// 						assert.Equal(t, expected[i].StringArrayValue, field.StringArrayValue)
		// 					}
		// 					if expected[i].StringSetValue.Status == pgtype.Present {
		// 						assert.Equal(t, expected[i].StringSetValue, field.StringSetValue)
		// 					}
		// 					if expected[i].IntSetValue.Status == pgtype.Present {
		// 						assert.Equal(t, expected[i].IntSetValue, field.IntSetValue)
		// 					}
		// 					assert.Equal(t, expected[i].FieldRenderGuide.Bytes, field.FieldRenderGuide.Bytes)
		// 				}
		// 			}).
		// 			Return(nil).
		// 			Once()
		// 	},
		// },
		{
			name: "submit new lesson report but there was current lesson report",
			lessonReport: &LessonReport{
				LessonID: "lesson-id-1",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
				LessonRepo:             lessonRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			},
			setup: func(ctx context.Context) {
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.LessonReport{
						LessonReportID: database.Text("lesson-report-id-1"),
						LessonID:       database.Text("lesson-id-1"),
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			err := tc.lessonReport.Submit(ctx, db)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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

func TestLessonReport_SaveDraft(t *testing.T) {
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
		name         string
		lessonReport *LessonReport
		setup        func(ctx context.Context)
		hasError     bool
	}{
		{
			name: "save new lesson report",
			lessonReport: &LessonReport{
				LessonID: "lesson-id-1",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
				LessonRepo:             lessonRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			},
			setup: func(ctx context.Context) {
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows)).Once()
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
					On("FindByPartnerAndFeatureName", ctx, db, database.Int4(1), database.Text(string(entities.FeatureNameIndividualLessonReport))).
					Return(&entities.PartnerFormConfig{
						FormConfigID:   database.Text("form-config-id-1"),
						PartnerID:      database.Int4(1),
						FeatureName:    database.Text(string(entities.FeatureNameIndividualLessonReport)),
						FormConfigData: database.JSONB(formConfigJson),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
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
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
			name: "save to update a lesson report",
			lessonReport: &LessonReport{
				LessonReportID: "lesson-report-id-1",
				LessonID:       "lesson-id-2", // lesson id is wrong, so it will be replaced by another id
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
				LessonRepo:             lessonRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			},
			setup: func(ctx context.Context) {
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}
				lessonReportRepo.
					On("FindByID", ctx, db, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSubmitted)), // report submitting status will be replaced
						FormConfigID:           database.Text("form-config-id-2"),                               // form config id will be replaced
					}, nil).Once()
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
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
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
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
					Return(nil).
					Once()
			},
		},
		{
			name: "submit new lesson report but there was current lesson report",
			lessonReport: &LessonReport{
				LessonID: "lesson-id-1",
				Details: LessonReportDetails{
					{
						StudentID:        "student-id-1",
						AttendanceStatus: entities.StudentAttendStatusAttend,
						AttendanceRemark: "very good",
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 5,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "title",
								Value: &AttributeValue{
									String: "monitor",
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "scores",
								Value: &AttributeValue{
									IntArray: []int{9, 10, 8, 10},
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "comments",
								Value: &AttributeValue{
									StringArray: []string{"excellent", "creative", "diligent"},
								},
							},
							{
								FieldID: "buddy",
								Value: &AttributeValue{
									StringSet: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"},
								},
							},
							{
								FieldID: "finished-exams",
								Value: &AttributeValue{
									IntSet: []int{1, 2, 3, 5, 6},
								},
							},
						},
					},
					{
						StudentID:        "student-id-2",
						AttendanceStatus: entities.StudentAttendStatusLeaveEarly,
						AttendanceNotice: entities.StudentAttendanceNoticeInAdvance,
						AttendanceReason: entities.StudentAttendanceReasonFamilyReason,
						AttendanceNote:   "lazy",
						Fields: LessonReportFields{
							{
								FieldID: "ordinal-number",
								Value: &AttributeValue{
									Int: 15,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
							{
								FieldID: "is-pass-lesson",
								Value: &AttributeValue{
									Bool: true,
								},
								FieldRenderGuide: []byte("fake guide to render this field"),
							},
						},
					},
				},
				LessonRepo:             lessonRepo,
				PartnerFormConfigRepo:  partnerFormConfigRepo,
				LessonReportRepo:       lessonReportRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonMemberRepo:       lessonMemberRepo,
				TeacherRepo:            teacherRepo,
			},
			setup: func(ctx context.Context) {
				report := &entities.LessonReport{}
				details := entities.LessonReportDetails{{}, {}}

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
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-id-1")).
					Return(database.TextArray([]string{"student-id-1", "student-id-2", "student-id-3"}), nil).
					Once()
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
				lessonReportRepo.
					On("FindByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.LessonReport{
						LessonReportID: database.Text("lesson-report-id-1"),
						LessonID:       database.Text("lesson-id-1"),
					}, nil).Once()
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
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
								AttendanceNotice: database.Text(string(entities.StudentAttendanceNoticeInAdvance)),
								AttendanceReason: database.Text(string(entities.StudentAttendanceReasonFamilyReason)),
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
					Return(nil).
					Once()
			},
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			err := tc.lessonReport.SaveDraft(ctx, db)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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

func TestLessonReport_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonReportRepo := new(mock_repositories.MockLessonReportRepo)
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)

	tcs := []struct {
		name           string
		LessonReportID string
		setup          func(ctx context.Context)
		hasError       bool
	}{
		{
			name:           "delete lesson report which have full data",
			LessonReportID: "lesson-report-id-1",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonReportRepo.
					On("FindByID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSubmitted)),
						FormConfigID:           database.Text("form-config-id-1"),
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(entities.LessonReportDetails{
						{
							LessonReportDetailID: database.Text("detail-id-1"),
							LessonReportID:       database.Text("lesson-report-id-1"),
							StudentID:            database.Text("student-id-1"),
							CreatedAt:            database.Timestamptz(now),
							UpdatedAt:            database.Timestamptz(now),
						},
						{
							LessonReportDetailID: database.Text("detail-id-2"),
							LessonReportID:       database.Text("lesson-report-id-1"),
							StudentID:            database.Text("student-id-2"),
							CreatedAt:            database.Timestamptz(now),
							UpdatedAt:            database.Timestamptz(now),
						},
					}, nil).
					Once()
				lessonMemberRepo.
					On("GetLessonMembersInLesson", ctx, tx, database.Text("lesson-id-1")).
					Return(entities.LessonMembers{
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-1"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
							AttendanceRemark: database.Text("very good"),
							CourseID:         database.Text("course-id-1"),
							UpdatedAt:        database.Timestamptz(now),
							CreatedAt:        database.Timestamptz(now),
						},
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-2"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text("very good"),
							CourseID:         database.Text("course-id-1"),
							UpdatedAt:        database.Timestamptz(now),
							CreatedAt:        database.Timestamptz(now),
						},
						{
							LessonID:  database.Text("lesson-id-1"),
							UserID:    database.Text("student-id-3"),
							CourseID:  database.Text("course-id-2"),
							UpdatedAt: database.Timestamptz(now),
							CreatedAt: database.Timestamptz(now),
						},
					}, nil).Once()
				lessonReportRepo.
					On("Delete", ctx, tx, database.Text("lesson-report-id-1")).
					Return(nil).Once()
				lessonReportDetailRepo.
					On("DeleteByLessonReportID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(nil).Once()
				lessonReportDetailRepo.
					On("DeleteFieldValuesByDetails", ctx, tx, database.TextArray([]string{"detail-id-1", "detail-id-2"})).
					Return(nil).Once()
				lessonMemberRepo.On("UpdateLessonMembersFields",
					ctx,
					tx,
					[]*entities.LessonMember{
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-1"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text(""),
						},
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-2"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text(""),
						},
					},
					entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
					},
				).Return(nil).Once()
			},
		},
		{
			name:           "case student id in report details not belong to lesson member",
			LessonReportID: "lesson-report-id-1",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonReportRepo.
					On("FindByID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSubmitted)),
						FormConfigID:           database.Text("form-config-id-1"),
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(entities.LessonReportDetails{
						{
							LessonReportDetailID: database.Text("detail-id-1"),
							StudentID:            database.Text("student-id-1"),
							LessonReportID:       database.Text("lesson-report-id-1"),
							CreatedAt:            database.Timestamptz(now),
							UpdatedAt:            database.Timestamptz(now),
						},
						{
							LessonReportDetailID: database.Text("detail-id-2"),
							StudentID:            database.Text("student-id-2"),
							LessonReportID:       database.Text("lesson-report-id-1"),
							CreatedAt:            database.Timestamptz(now),
							UpdatedAt:            database.Timestamptz(now),
						},
						{
							LessonReportDetailID: database.Text("detail-id-3"),
							StudentID:            database.Text("student-id-5"),
							LessonReportID:       database.Text("lesson-report-id-1"),
							CreatedAt:            database.Timestamptz(now),
							UpdatedAt:            database.Timestamptz(now),
						},
					}, nil).
					Once()
				lessonMemberRepo.
					On("GetLessonMembersInLesson", ctx, tx, database.Text("lesson-id-1")).
					Return(entities.LessonMembers{
						{
							UserID:           database.Text("student-id-1"),
							LessonID:         database.Text("lesson-id-1"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
							AttendanceRemark: database.Text("very good"),
							CourseID:         database.Text("course-id-1"),
							UpdatedAt:        database.Timestamptz(now),
							CreatedAt:        database.Timestamptz(now),
						},
						{
							UserID:           database.Text("student-id-2"),
							LessonID:         database.Text("lesson-id-1"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text("very good"),
							CourseID:         database.Text("course-id-1"),
							UpdatedAt:        database.Timestamptz(now),
							CreatedAt:        database.Timestamptz(now),
						},
						{
							UserID:    database.Text("student-id-3"),
							LessonID:  database.Text("lesson-id-1"),
							CourseID:  database.Text("course-id-2"),
							UpdatedAt: database.Timestamptz(now),
							CreatedAt: database.Timestamptz(now),
						},
					}, nil).Once()
				lessonReportRepo.
					On("Delete", ctx, tx, database.Text("lesson-report-id-1")).
					Return(nil).Once()
				lessonReportDetailRepo.
					On("DeleteByLessonReportID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(nil).Once()
				lessonReportDetailRepo.
					On("DeleteFieldValuesByDetails",
						ctx,
						tx,
						database.TextArray([]string{"detail-id-1", "detail-id-2", "detail-id-3"})).
					Return(nil).Once()
				lessonMemberRepo.On("UpdateLessonMembersFields",
					ctx,
					tx,
					[]*entities.LessonMember{
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-1"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text(""),
						},
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-2"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text(""),
						},
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-5"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text(""),
						},
					},
					entities.UpdateLessonMemberFields{
						entities.LessonMemberAttendanceRemark,
						entities.LessonMemberAttendanceStatus,
					},
				).Return(nil).Once()
			},
		},
		{
			name:           "delete none exist lesson report",
			LessonReportID: "lesson-report-id-1",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonReportRepo.
					On("FindByID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(nil, pgx.ErrNoRows).
					Once()
			},
			hasError: true,
		},
		{
			name:           "delete lesson report without details",
			LessonReportID: "lesson-report-id-1",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonReportRepo.
					On("FindByID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(&entities.LessonReport{
						LessonReportID:         database.Text("lesson-report-id-1"),
						LessonID:               database.Text("lesson-id-1"),
						ReportSubmittingStatus: database.Text(string(entities.ReportSubmittingStatusSubmitted)),
						FormConfigID:           database.Text("form-config-id-1"),
					}, nil).
					Once()
				lessonReportDetailRepo.
					On("GetByLessonReportID", ctx, tx, database.Text("lesson-report-id-1")).
					Return(entities.LessonReportDetails{}, nil).
					Once()
				lessonMemberRepo.
					On("GetLessonMembersInLesson", ctx, tx, database.Text("lesson-id-1")).
					Return(entities.LessonMembers{
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-1"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusAttend)),
							AttendanceRemark: database.Text("very good"),
							CourseID:         database.Text("course-id-1"),
							UpdatedAt:        database.Timestamptz(now),
							CreatedAt:        database.Timestamptz(now),
						},
						{
							LessonID:         database.Text("lesson-id-1"),
							UserID:           database.Text("student-id-2"),
							AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
							AttendanceRemark: database.Text("very good"),
							CourseID:         database.Text("course-id-1"),
							UpdatedAt:        database.Timestamptz(now),
							CreatedAt:        database.Timestamptz(now),
						},
						{
							LessonID:  database.Text("lesson-id-1"),
							UserID:    database.Text("student-id-3"),
							CourseID:  database.Text("course-id-2"),
							UpdatedAt: database.Timestamptz(now),
							CreatedAt: database.Timestamptz(now),
						},
					}, nil).Once()
				lessonReportRepo.
					On("Delete", ctx, tx, database.Text("lesson-report-id-1")).
					Return(nil).Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			lr := &LessonReport{
				LessonReportID:         tc.LessonReportID,
				LessonMemberRepo:       lessonMemberRepo,
				LessonReportDetailRepo: lessonReportDetailRepo,
				LessonReportRepo:       lessonReportRepo,
			}
			err := lr.Delete(ctx, db)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				lessonReportRepo,
				lessonReportDetailRepo,
				lessonMemberRepo,
			)
		})
	}
}
