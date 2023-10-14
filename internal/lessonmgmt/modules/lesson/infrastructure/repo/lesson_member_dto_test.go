package repo

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLessonMembersFromLessonEntity(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		lesson  *domain.Lesson
		members LessonMembers
	}{
		{
			name: "full fields",
			lesson: &domain.Lesson{
				LessonID: "lesson-id-1",
				Learners: domain.LessonLearners{
					{
						LearnerID:        "leaner-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.FamilyReason,
					},
					{
						LearnerID:        "leaner-id-2",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						AttendanceNotice: domain.OnTheDay,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
					},
					{
						LearnerID:        "leaner-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusEmpty,
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
				},
			},
			members: LessonMembers{
				{
					LessonID:         database.Text("lesson-id-1"),
					UserID:           database.Text("leaner-id-1"),
					AttendanceStatus: database.Text(string(domain.StudentAttendStatusAbsent)),
					AttendanceRemark: pgtype.Text{Status: pgtype.Null},
					CourseID:         database.Text("course-id-1"),
					AttendanceNotice: database.Text(string(domain.InAdvance)),
					AttendanceReason: database.Text(string(domain.FamilyReason)),
					AttendanceNote:   pgtype.Text{Status: pgtype.Null},
					UpdatedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					CreatedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					DeletedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					UserFirstName:    pgtype.Text{Status: pgtype.Null},
					UserLastName:     pgtype.Text{Status: pgtype.Null},
				},
				{
					LessonID:         database.Text("lesson-id-1"),
					UserID:           database.Text("leaner-id-2"),
					AttendanceStatus: database.Text(string(domain.StudentAttendStatusAbsent)),
					AttendanceRemark: pgtype.Text{Status: pgtype.Null},
					CourseID:         database.Text("course-id-1"),
					AttendanceNotice: database.Text(string(domain.OnTheDay)),
					AttendanceReason: database.Text(string(domain.PhysicalCondition)),
					AttendanceNote:   database.Text("sample-attendance-note"),
					UpdatedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					CreatedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					DeletedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					UserFirstName:    pgtype.Text{Status: pgtype.Null},
					UserLastName:     pgtype.Text{Status: pgtype.Null},
				},
				{
					LessonID:         database.Text("lesson-id-1"),
					UserID:           database.Text("leaner-id-3"),
					AttendanceStatus: database.Text(string(domain.StudentAttendStatusEmpty)),
					AttendanceRemark: pgtype.Text{Status: pgtype.Null},
					CourseID:         database.Text("course-id-3"),
					AttendanceNotice: database.Text(string(domain.NoticeEmpty)),
					AttendanceReason: database.Text(string(domain.ReasonEmpty)),
					AttendanceNote:   pgtype.Text{Status: pgtype.Null},
					UpdatedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					CreatedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					DeletedAt:        pgtype.Timestamptz{Status: pgtype.Null},
					UserFirstName:    pgtype.Text{Status: pgtype.Null},
					UserLastName:     pgtype.Text{Status: pgtype.Null},
				},
			},
		},
		{
			name: "there are no any memeber",
			lesson: &domain.Lesson{
				LessonID: "lesson-id-1",
			},
			members: LessonMembers{},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLessonMembersFromLessonEntity(tc.lesson)
			require.NoError(t, err)
			assert.EqualValues(t, tc.members, actual)
		})
	}
}
func TestStringArray(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                string
		expectedStringArray []string
	}{
		{
			name:                "full fields",
			expectedStringArray: []string{"test-1", "test-2"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var updateLessonMemberFields = UpdateLessonMemberFields{
				"test-1", "test-2",
			}
			actual := updateLessonMemberFields.StringArray()

			assert.EqualValues(t, tc.expectedStringArray, actual)
		})
	}
}
