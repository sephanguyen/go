package domain

import (
	"testing"
	"time"

	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"

	"github.com/stretchr/testify/assert"
)

var now = time.Now()

func TestLessonReportDetails_ToLessonMembersEntity(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                        string
		expectedLessonMemberDomains lesson_domain.LessonMembers
	}{
		{
			name: "full fields",
			expectedLessonMemberDomains: lesson_domain.LessonMembers{
				&lesson_domain.LessonMember{
					LessonID:         "test-lesson-id-1",
					StudentID:        "test-student-id-1",
					AttendanceStatus: "test-status",
					AttendanceRemark: "test-remark",
					AttendanceReason: "test-reason",
					AttendanceNotice: "test-notice",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
				&lesson_domain.LessonMember{
					LessonID:         "test-lesson-id-1",
					StudentID:        "test-student-id-2",
					AttendanceStatus: "test-status",
					AttendanceRemark: "test-remark",
					AttendanceReason: "test-reason",
					AttendanceNotice: "test-notice",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var lessonReportDetails = LessonReportDetails{
				&LessonReportDetail{
					StudentID:        "test-student-id-1",
					AttendanceStatus: "test-status",
					AttendanceRemark: "test-remark",
					AttendanceReason: "test-reason",
					AttendanceNotice: "test-notice",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
				&LessonReportDetail{
					StudentID:        "test-student-id-2",
					AttendanceStatus: "test-status",
					AttendanceRemark: "test-remark",
					AttendanceReason: "test-reason",
					AttendanceNotice: "test-notice",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			}
			actual := lessonReportDetails.ToLessonMembersEntity("test-lesson-id-1")
			//ignore testing default timestamp
			for i := 0; i < len(actual); i++ {
				actual[i].CreatedAt = now
				actual[i].UpdatedAt = now
			}

			assert.EqualValues(t, tc.expectedLessonMemberDomains, actual)
		})
	}
}

func TestLessonReportDetails_ToLessonReportDetailsDomain(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                        string
		expectedLessonReportDomains LessonReportDetails
		isCreateLessonReport        bool
	}{
		{
			name: "full fields",
			expectedLessonReportDomains: LessonReportDetails{
				&LessonReportDetail{
					LessonReportDetailID: "test-report-id-1",
					LessonReportID:       "test-report-id-1",
					StudentID:            "test-student-id-1",
					CreatedAt:            now,
					UpdatedAt:            now,
					ReportVersion:        1,
				},
				&LessonReportDetail{
					LessonReportDetailID: "test-report-id-1",
					LessonReportID:       "test-report-id-1",
					StudentID:            "test-student-id-2",
					CreatedAt:            now,
					UpdatedAt:            now,
					ReportVersion:        1,
				},
			},
			isCreateLessonReport: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var lessonReportDetails = LessonReportDetails{
				&LessonReportDetail{
					StudentID:     "test-student-id-1",
					CreatedAt:     now,
					UpdatedAt:     now,
					ReportVersion: 1,
				},
				&LessonReportDetail{
					StudentID:     "test-student-id-2",
					CreatedAt:     now,
					UpdatedAt:     now,
					ReportVersion: 0,
				},
			}
			actual, err := lessonReportDetails.ToLessonReportDetailsDomain("test-report-id-1")
			//ignore testing default timestamp
			for i := 0; i < len(actual); i++ {
				actual[i].LessonReportDetailID = "test-report-id-1"
				actual[i].CreatedAt = now
				actual[i].UpdatedAt = now
			}
			assert.Nil(t, err)
			assert.EqualValues(t, tc.expectedLessonReportDomains, actual)
		})
	}
}

func TestLessonReportDetails_RemoveAttendanceInfo(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                        string
		expectedLessonReportDomains LessonReportDetails
	}{
		{
			name: "full fields",
			expectedLessonReportDomains: LessonReportDetails{
				&LessonReportDetail{
					LessonReportDetailID: "test-report-id-1",
					LessonReportID:       "test-report-id-1",
					StudentID:            "test-student-id-1",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
				&LessonReportDetail{
					LessonReportDetailID: "test-report-id-1",
					LessonReportID:       "test-report-id-1",
					StudentID:            "test-student-id-2",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var lessonReportDetails = LessonReportDetails{
				&LessonReportDetail{
					LessonReportDetailID: "test-report-id-1",
					LessonReportID:       "test-report-id-1",
					StudentID:            "test-student-id-1",
					CreatedAt:            now,
					UpdatedAt:            now,
					AttendanceStatus:     constant.StudentAttendStatusAbsent,
					AttendanceRemark:     "remark",
					AttendanceNotice:     constant.StudentAttendanceNoticeInAdvance,
					AttendanceReason:     constant.StudentAttendanceReasonEmpty,
					AttendanceNote:       "",
				},
				&LessonReportDetail{
					LessonReportDetailID: "test-report-id-1",
					LessonReportID:       "test-report-id-1",
					StudentID:            "test-student-id-2",
					CreatedAt:            now,
					UpdatedAt:            now,
					AttendanceStatus:     constant.StudentAttendStatusAbsent,
					AttendanceRemark:     "remark",
					AttendanceNotice:     constant.StudentAttendanceNoticeInAdvance,
					AttendanceReason:     constant.StudentAttendanceReasonEmpty,
					AttendanceNote:       "",
				},
			}
			err := lessonReportDetails.RemoveAttendanceInfo()
			//ignore testing default timestamp
			for i := 0; i < len(lessonReportDetails); i++ {
				lessonReportDetails[i].LessonReportDetailID = "test-report-id-1"
				lessonReportDetails[i].CreatedAt = now
				lessonReportDetails[i].UpdatedAt = now
			}
			assert.Nil(t, err)
			assert.EqualValues(t, tc.expectedLessonReportDomains, lessonReportDetails)
		})
	}
}
