package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLesson_Normalize(t *testing.T) {
	t.Parallel()
	now := time.Time{}

	tcs := []struct {
		name     string
		lesson   *VirtualLesson
		expected *VirtualLesson
	}{
		{
			name: "missing Status, LessonType, TeachingModel, StreamLearnerCounter, LearnerIds field",
			lesson: &VirtualLesson{
				LessonID:             database.Text("lesson-id-1"),
				Name:                 database.Text("name 1"),
				TeacherID:            database.Text("teacher-id-1"),
				CourseID:             database.Text("course-id-1"),
				ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
				CreatedAt:            database.Timestamptz(now),
				UpdatedAt:            database.Timestamptz(now),
				DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
				EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
				StartTime:            database.Timestamptz(now.Add(2 * time.Minute)),
				EndTime:              database.Timestamptz(now.Add(20 * time.Minute)),
				LessonGroupID:        database.Text("lesson-gr-id-1"),
				RoomID:               database.Text("room-id-2"),
				StreamLearnerCounter: pgtype.Int4{Status: pgtype.Null},
				RoomState:            pgtype.JSONB{Status: pgtype.Null},
				ClassID:              database.Text("class-id-2"),
				CenterID:             database.Text("center-id-1"),
				TeachingMedium:       database.Text(string(domain.LessonTeachingMediumOnline)),
				TeachingMethod:       database.Text(string(domain.LessonTeachingMethodIndividual)),
				SchedulingStatus:     database.Text(string(domain.LessonSchedulingStatusPublished)),
				LearnerIDs:           LearnerIDs{LearnerIDs: database.TextArray([]string{"learner-id-1"})},
				TeacherIDs:           TeacherIDs{TeacherIDs: database.TextArray([]string{"teacher-id-1"})},
				CourseIDs:            CourseIDs{CourseIDs: database.TextArray([]string{"course-id-1"})},
			},
			expected: &VirtualLesson{
				LessonID:             database.Text("lesson-id-1"),
				Name:                 database.Text("name 1"),
				TeacherID:            database.Text("teacher-id-1"),
				CourseID:             database.Text("course-id-1"),
				ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
				CreatedAt:            database.Timestamptz(now),
				UpdatedAt:            database.Timestamptz(now),
				DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
				EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
				StartTime:            database.Timestamptz(now.Add(2 * time.Minute)),
				EndTime:              database.Timestamptz(now.Add(20 * time.Minute)),
				LessonGroupID:        database.Text("lesson-gr-id-1"),
				RoomID:               database.Text("room-id-2"),
				StreamLearnerCounter: database.Int4(0),
				RoomState:            pgtype.JSONB{Status: pgtype.Null},
				ClassID:              database.Text("class-id-2"),
				CenterID:             database.Text("center-id-1"),
				TeachingMedium:       database.Text(string(domain.LessonTeachingMediumOnline)),
				TeachingMethod:       database.Text(string(domain.LessonTeachingMethodIndividual)),
				SchedulingStatus:     database.Text(string(domain.LessonSchedulingStatusPublished)),
				LearnerIDs:           LearnerIDs{LearnerIDs: database.TextArray([]string{"learner-id-1"})},
				TeacherIDs:           TeacherIDs{TeacherIDs: database.TextArray([]string{"teacher-id-1"})},
				CourseIDs:            CourseIDs{CourseIDs: database.TextArray([]string{"course-id-1"})},
			},
		},
		{
			name: "full fields",
			lesson: &VirtualLesson{
				LessonID:             database.Text("lesson-id-1"),
				Name:                 database.Text("name 1"),
				TeacherID:            database.Text("teacher-id-1"),
				CourseID:             database.Text("course-id-1"),
				ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
				CreatedAt:            database.Timestamptz(now),
				UpdatedAt:            database.Timestamptz(now),
				DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
				EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
				StartTime:            database.Timestamptz(now.Add(2 * time.Minute)),
				EndTime:              database.Timestamptz(now.Add(20 * time.Minute)),
				LessonGroupID:        database.Text("lesson-gr-id-1"),
				RoomID:               database.Text("room-id-2"),
				StreamLearnerCounter: database.Int4(2),
				RoomState:            pgtype.JSONB{Status: pgtype.Null},
				ClassID:              database.Text("class-id-2"),
				CenterID:             database.Text("center-id-1"),
				TeachingMedium:       database.Text(string(domain.LessonTeachingMediumOnline)),
				TeachingMethod:       database.Text(string(domain.LessonTeachingMethodIndividual)),
				SchedulingStatus:     database.Text(string(domain.LessonSchedulingStatusPublished)),
				LearnerIDs:           LearnerIDs{LearnerIDs: database.TextArray([]string{"learner-id-1"})},
				TeacherIDs:           TeacherIDs{TeacherIDs: database.TextArray([]string{"teacher-id-1"})},
				CourseIDs:            CourseIDs{CourseIDs: database.TextArray([]string{"course-id-1"})},
			},
			expected: &VirtualLesson{
				LessonID:             database.Text("lesson-id-1"),
				Name:                 database.Text("name 1"),
				TeacherID:            database.Text("teacher-id-1"),
				CourseID:             database.Text("course-id-1"),
				ControlSettings:      pgtype.JSONB{Status: pgtype.Null},
				CreatedAt:            database.Timestamptz(now),
				UpdatedAt:            database.Timestamptz(now),
				DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
				EndAt:                pgtype.Timestamptz{Status: pgtype.Null},
				StartTime:            database.Timestamptz(now.Add(2 * time.Minute)),
				EndTime:              database.Timestamptz(now.Add(20 * time.Minute)),
				LessonGroupID:        database.Text("lesson-gr-id-1"),
				RoomID:               database.Text("room-id-2"),
				StreamLearnerCounter: database.Int4(2),
				RoomState:            pgtype.JSONB{Status: pgtype.Null},
				ClassID:              database.Text("class-id-2"),
				CenterID:             database.Text("center-id-1"),
				TeachingMedium:       database.Text(string(domain.LessonTeachingMediumOnline)),
				TeachingMethod:       database.Text(string(domain.LessonTeachingMethodIndividual)),
				SchedulingStatus:     database.Text(string(domain.LessonSchedulingStatusPublished)),
				LearnerIDs:           LearnerIDs{LearnerIDs: database.TextArray([]string{"learner-id-1"})},
				TeacherIDs:           TeacherIDs{TeacherIDs: database.TextArray([]string{"teacher-id-1"})},
				CourseIDs:            CourseIDs{CourseIDs: database.TextArray([]string{"course-id-1"})},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.lesson.Normalize()
			require.NoError(t, err)
			assert.EqualValues(t, tc.expected, tc.lesson)
		})
	}
}
