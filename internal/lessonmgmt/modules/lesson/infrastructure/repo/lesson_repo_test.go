package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonRepoWithSqlMock() (*LessonRepo, *testutil.MockDB) {
	r := &LessonRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonRepo_InsertLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	r, mockDB := LessonRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		lesson := &domain.Lesson{
			LessonID:         "",
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			SchedulerID: "scheduler-id",
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
			PreparationTime: 120,
			BreakTime:       30,
		}

		// mock insert lesson gr
		gr := &LessonGroup{
			CourseID: database.Text("course-id-1"),
			MediaIDs: database.TextArray([]string{"media-id-1", "media-id-2"}),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock insert lesson
		lessonDto, _ := NewLessonFromEntity(lesson)
		_ = lessonDto.Normalize()
		args = append([]interface{}{
			mock.Anything, mock.AnythingOfType("string"),
		},
			mock.Anything, &lessonDto.TeacherID, &lessonDto.CourseID,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, &lessonDto.LessonType, &lessonDto.Status,
			mock.Anything, mock.Anything, &lessonDto.Name, mock.Anything, mock.Anything,
			mock.Anything, &lessonDto.TeachingModel, mock.Anything, &lessonDto.CenterID,
			&lessonDto.TeachingMedium, &lessonDto.TeachingMethod, &lessonDto.SchedulingStatus, &lessonDto.SchedulerID, &lessonDto.IsLocked, mock.Anything, mock.Anything, mock.Anything, mock.Anything, &lessonDto.LessonCapacity,
			&lessonDto.PreparationTime, &lessonDto.BreakTime, mock.Anything, mock.Anything, mock.Anything,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock upsert teachers
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Teachers)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Classrooms)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert lesson course
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		lesson, err := r.InsertLesson(ctx, mockDB.DB, lesson)
		require.NoError(t, err)
		require.False(t, lesson.CreatedAt.IsZero())
		require.False(t, lesson.UpdatedAt.IsZero())

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			batchResults,
		)
	})

	t.Run("got error", func(t *testing.T) {

		lesson := &domain.Lesson{
			LessonID:         "",
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			SchedulerID: "scheduler-id",
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
			PreparationTime: 150,
			BreakTime:       10,
		}

		// mock insert lesson gr
		gr := &LessonGroup{
			CourseID: database.Text("course-id-1"),
			MediaIDs: database.TextArray([]string{"media-id-1", "media-id-2"}),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock insert lesson
		lessonDto, _ := NewLessonFromEntity(lesson)
		_ = lessonDto.Normalize()
		args = append([]interface{}{
			mock.Anything, mock.AnythingOfType("string"),
		},
			mock.Anything, &lessonDto.TeacherID, &lessonDto.CourseID,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, &lessonDto.LessonType, &lessonDto.Status,
			mock.Anything, mock.Anything, &lessonDto.Name, mock.Anything, mock.Anything,
			mock.Anything, &lessonDto.TeachingModel, mock.Anything, &lessonDto.CenterID,
			&lessonDto.TeachingMedium, &lessonDto.TeachingMethod, &lessonDto.SchedulingStatus, &lessonDto.SchedulerID, &lessonDto.IsLocked, mock.Anything, mock.Anything, mock.Anything, mock.Anything, &lessonDto.LessonCapacity,
			&lessonDto.PreparationTime, &lessonDto.BreakTime, mock.Anything, mock.Anything, mock.Anything,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock upsert teachers
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Teachers)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Classrooms)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert lesson course
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), errors.New("error")).Once()
		batchResults.On("Close").Once().Return(nil)

		lesson, err := r.InsertLesson(ctx, mockDB.DB, lesson)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			batchResults,
		)
	})
}

func TestLessonRepo_UpdateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	r, mockDB := LessonRepoWithSqlMock()
	t.Run("successfully with creating new lesson group", func(t *testing.T) {
		lesson := &domain.Lesson{
			LessonID:         "lesson-id-1",
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			SchedulerID: "scheduler-id-1",
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
		}

		// mock get current lesson by id
		mockDB.MockQueryRowArgs(
			t,
			[]interface{}{mock.Anything, mock.AnythingOfType("string"), &lesson.LessonID}...,
		)
		currentLessonDto := &Lesson{
			CourseID:      database.Text("course-id-2"),
			LessonGroupID: database.Text("lesson-group-id-1"),
		}
		fields, values := currentLessonDto.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)

		// mock insert lesson gr
		gr := &LessonGroup{
			CourseID: database.Text("course-id-1"),
			MediaIDs: database.TextArray([]string{"media-id-1", "media-id-2"}),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock update lesson
		lessonDto, _ := NewLessonFromEntity(lesson)
		_ = lessonDto.Normalize()
		args = append([]interface{}{
			mock.Anything, mock.AnythingOfType("string"),
		},
			&lessonDto.TeacherID, &lessonDto.CourseID,
			mock.Anything, mock.Anything, // "updated_at", "lesson_group_id"
			&lessonDto.LessonType, &lessonDto.Status, // "lesson_type", "status"
			mock.Anything, mock.Anything, &lessonDto.TeachingModel, // "start_time", "end_time", "teaching_model"
			&lessonDto.CenterID, &lessonDto.TeachingMedium, &lessonDto.TeachingMethod,
			&lessonDto.SchedulingStatus, &lessonDto.ClassID, &lessonDto.SchedulerID,
			&lessonDto.ZoomLink, &lessonDto.ZoomOwnerID, &lessonDto.ZoomID,
			&lessonDto.LessonCapacity, &lessonDto.ClassDoOwnerID, &lessonDto.ClassDoLink, &lessonDto.ClassDoRoomID,
			&lessonDto.PreparationTime, &lessonDto.BreakTime, &lessonDto.LessonID,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		// mock upsert teachers
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Teachers)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Classrooms)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert lesson course
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		lesson, err := r.UpdateLesson(ctx, mockDB.DB, lesson)
		require.NoError(t, err)
		require.False(t, lesson.CreatedAt.IsZero())
		require.False(t, lesson.UpdatedAt.IsZero())

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			batchResults,
		)
	})

	t.Run("successfully with updating current lesson group", func(t *testing.T) {
		lesson := &domain.Lesson{
			LessonID:         "lesson-id-1",
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			SchedulerID: "scheduler-id-1",
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
		}

		// mock get current lesson by id
		mockDB.MockQueryRowArgs(
			t,
			[]interface{}{mock.Anything, mock.AnythingOfType("string"), &lesson.LessonID}...,
		)
		currentLessonDto := &Lesson{
			CourseID:      database.Text("course-id-1"), // same course so will update
			LessonGroupID: database.Text("lesson-group-id-1"),
		}
		fields, values := currentLessonDto.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)

		// mock update media for lesson gr
		gr := &LessonGroup{
			CourseID: database.Text("course-id-1"),
			MediaIDs: database.TextArray([]string{"media-id-1", "media-id-2"}),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &gr.MediaIDs, mock.Anything, &currentLessonDto.LessonGroupID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock update lesson
		lessonDto, _ := NewLessonFromEntity(lesson)
		_ = lessonDto.Normalize()
		args = append([]interface{}{
			mock.Anything, mock.AnythingOfType("string"),
		},
			&lessonDto.TeacherID, &lessonDto.CourseID,
			mock.Anything, &currentLessonDto.LessonGroupID, // "updated_at", "lesson_group_id"
			&lessonDto.LessonType, &lessonDto.Status, // "lesson_type", "status"
			mock.Anything, mock.Anything, &lessonDto.TeachingModel, // "start_time", "end_time", "teaching_model"
			&lessonDto.CenterID, &lessonDto.TeachingMedium, &lessonDto.TeachingMethod,
			&lessonDto.SchedulingStatus, &lessonDto.ClassID, &lessonDto.SchedulerID, &lessonDto.ZoomLink, &lessonDto.ZoomOwnerID, &lessonDto.ZoomID,
			&lessonDto.LessonCapacity, &lessonDto.ClassDoOwnerID, &lessonDto.ClassDoLink, &lessonDto.ClassDoRoomID,
			&lessonDto.PreparationTime, &lessonDto.BreakTime, &lessonDto.LessonID,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock upsert teachers
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Teachers)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Classrooms)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert lesson course
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		lesson, err := r.UpdateLesson(ctx, mockDB.DB, lesson)
		require.NoError(t, err)
		require.False(t, lesson.CreatedAt.IsZero())
		require.False(t, lesson.UpdatedAt.IsZero())

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			batchResults,
		)
	})

	t.Run("got error", func(t *testing.T) {
		lesson := &domain.Lesson{
			LessonID:         "lesson-id-1",
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
		}

		// mock get current lesson by id
		mockDB.MockQueryRowArgs(
			t,
			[]interface{}{mock.Anything, mock.AnythingOfType("string"), &lesson.LessonID}...,
		)
		currentLessonDto := &Lesson{
			CourseID:      database.Text("course-id-2"),
			LessonGroupID: database.Text("lesson-group-id-1"),
		}
		fields, values := currentLessonDto.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)

		// mock insert lesson gr
		gr := &LessonGroup{
			CourseID: database.Text("course-id-1"),
			MediaIDs: database.TextArray([]string{"media-id-1", "media-id-2"}),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock update lesson
		lessonDto, _ := NewLessonFromEntity(lesson)
		_ = lessonDto.Normalize()
		args = append([]interface{}{
			mock.Anything, mock.AnythingOfType("string"),
		},
			&lessonDto.TeacherID, &lessonDto.CourseID,
			mock.Anything, mock.Anything, // "updated_at", "lesson_group_id"
			&lessonDto.LessonType, &lessonDto.Status, // "lesson_type", "status"
			mock.Anything, mock.Anything, &lessonDto.TeachingModel, // "start_time", "end_time", "teaching_model"
			&lessonDto.CenterID, &lessonDto.TeachingMedium, &lessonDto.TeachingMethod,
			&lessonDto.SchedulingStatus, &lessonDto.ClassID, &lessonDto.SchedulerID, &lessonDto.ZoomLink, &lessonDto.ZoomOwnerID, &lessonDto.ZoomID,
			&lessonDto.LessonCapacity, &lessonDto.ClassDoOwnerID, &lessonDto.ClassDoLink, &lessonDto.ClassDoRoomID,
			&lessonDto.PreparationTime, &lessonDto.BreakTime, &lessonDto.LessonID,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock upsert teachers
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Teachers)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Classrooms)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(lesson.Learners)+1; i++ {
			batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		// mock upsert lesson course
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), errors.New("error")).Once()
		batchResults.On("Close").Once().Return(nil)

		lesson, err := r.UpdateLesson(ctx, mockDB.DB, lesson)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			batchResults,
		)
	})
}

func TestLessonRepo_GetLessonByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lessonIDs := []string{"lesson-1", "lesson-2"}
	lesson := &Lesson{}
	lessonFields, lessonValues := lesson.FieldMap()
	lessonGroup := &LessonGroup{}
	lessonGroupFields, lessonGroupValues := lessonGroup.FieldMap()
	lessonTeacher := &LessonTeacher{}
	lessonTeacherFields, lessonTeacherValues := lessonTeacher.FieldMap()
	lessonMemberFields, lessonMemberValues := (&entities.LessonMember{}).FieldMap()

	t.Run("success", func(t *testing.T) {
		//lesson
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, lessonFields, lessonValues)

		//lesson-teacher
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, lessonTeacherFields, lessonTeacherValues)

		// //lesson-learner
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, lessonMemberFields, [][]interface{}{
			lessonMemberValues,
		})
		gotLessons, err := r.GetLessonByIDs(ctx, mockDB.DB, lessonIDs)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessons)
	})

	t.Run("error get lesson", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonIDs)
		gotLessonMembers, err := r.GetLessonByIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLessonMembers)
	})

	t.Run("error get teacher", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonIDs)
		mockDB.MockScanFields(nil, lessonFields, lessonValues)

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, lessonGroupFields, lessonGroupValues)

		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything)
		gotLessonMembers, err := r.GetLessonByIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLessonMembers)
	})
}

func TestLessonRepo_LockLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	lessonIds := []string{"lesson-1", "lesson-2", "lesson-3"}

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.LockLesson(ctx, mockDB.DB, lessonIds)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything)
		for i := 0; i < len(lessonIds); i++ {
			mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		}

		err := r.LockLesson(ctx, mockDB.DB, lessonIds)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "lessons")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "is_locked", "updated_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
		})
	})
}

func TestLessonRepo_UpdateLessonSchedulingStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	lesson := &domain.Lesson{
		LessonID:         "lesson-id-1",
		Name:             "lesson name",
		LocationID:       "center id",
		SchedulingStatus: domain.LessonSchedulingStatusCanceled,
	}

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		_, err := r.UpdateLessonSchedulingStatus(ctx, mockDB.DB, lesson)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		_, err := r.UpdateLessonSchedulingStatus(ctx, mockDB.DB, lesson)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "lessons")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "scheduling_status", "updated_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 3}},
		})
	})
}

func TestLessonRepo_UpdateSchedulerID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	lessonIds := []string{"lesson-1", "lesson-2", "lesson-3"}
	schedulerId := "scheduler-id"

	t.Run("err update", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), errors.New("internal error")).Times(len(lessonIds))

		batchResults.On("Close").Once().Return(nil)

		err := r.UpdateSchedulerID(ctx, mockDB.DB, lessonIds, schedulerId)
		assert.Error(t, err, "internal error")
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Times(len(lessonIds))

		batchResults.On("Close").Once().Return(nil)
		err := r.UpdateSchedulerID(ctx, mockDB.DB, lessonIds, schedulerId)
		assert.Nil(t, err)
	})
}

func TestLessonRepo_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	lessonIDs := []string{"lesson-id"}

	t.Run("err delete", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &lessonIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.Delete(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &lessonIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Delete(ctx, mockDB.DB, lessonIDs)
		assert.Nil(t, err)
	})
}

func TestLessonRepo_GetLessonRecurring(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	lessonID := "lesson-id"
	schedulerID := "scheduler-id"

	t.Run("err delete recurring due to cannot get lessonID", func(t *testing.T) {

		mockDB.MockQueryRowArgs(
			t,
			[]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID}...,
		)
		currentLessonDto := &Lesson{
			StartTime: database.Timestamptz(time.Now()),
		}
		fields, values := currentLessonDto.FieldMap()

		mockDB.MockRowScanFields(fmt.Errorf("could not get current lesson: %s", lessonID), fields, values)

		_, err := r.GetFutureRecurringLessonIDs(ctx, mockDB.DB, lessonID)
		assert.EqualError(t, err, "could not get current lesson: "+lessonID)
	})

	t.Run("err delete recurring", func(t *testing.T) {
		currentLessonDto := &Lesson{
			StartTime:   database.Timestamptz(time.Now()),
			SchedulerID: database.Text(schedulerID),
		}
		fields, values := currentLessonDto.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		mockDB.MockQueryRowArgs(
			t,
			[]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID}...,
		)

		mockDB.DB.On("Query", mock.Anything, mock.Anything, &currentLessonDto.SchedulerID, &currentLessonDto.StartTime).Return(nil, puddle.ErrClosedPool)

		_, err := r.GetFutureRecurringLessonIDs(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// mock get current lesson
		currentLessonDto := &Lesson{
			StartTime:   database.Timestamptz(time.Now()),
			SchedulerID: database.Text(schedulerID),
		}
		fields, values := currentLessonDto.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		mockDB.MockQueryRowArgs(
			t,
			[]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID}...,
		)
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &currentLessonDto.SchedulerID, &currentLessonDto.StartTime).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Close").Once().Return(nil)
		mockDB.Rows.On("Next").Once().Return(true)
		var l pgtype.Text
		mockDB.Rows.On("Scan", &l).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)

		_, err := r.GetFutureRecurringLessonIDs(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
	})
}

func TestLessonRepo_UpsertLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	r, mockDB := LessonRepoWithSqlMock()
	t.Run("success", func(t *testing.T) {
		baseLesson := &domain.Lesson{
			LessonID:         idutil.ULIDNow(),
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
		}
		lessons := []*domain.Lesson{baseLesson}
		for i := 1; i < 5; i++ {
			lesson := *baseLesson
			lesson.LessonID = idutil.ULIDNow()
			lesson.StartTime = lesson.StartTime.AddDate(0, 0, 7)
			lesson.EndTime = lesson.StartTime.AddDate(0, 0, 7)
			learners := domain.LessonLearners{}
			for _, learner := range lesson.Learners {
				l := *learner
				l.AttendStatus = domain.StudentAttendStatusEmpty
				learners = append(learners, &l)
			}
			lesson.Learners = learners
			lesson.Material = &domain.LessonMaterial{}
			lessons = append(lessons, &lesson)
		}
		recurringLesson := &domain.RecurringLesson{
			Lessons: lessons,
		}
		for _, lesson := range lessons {
			// mock insert lesson gr
			gr := &LessonGroup{
				CourseID: database.Text("course-id-1"),
				MediaIDs: pgtype.TextArray{Status: pgtype.Null},
			}
			if len(lesson.Material.MediaIDs) != 0 {
				gr.MediaIDs = database.TextArray([]string{"media-id-1", "media-id-2"})
			}
			args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")},
				mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
			mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		}
		// mock insert lesson
		lessonsDto := make([]*Lesson, 0)
		for _, lesson := range recurringLesson.Lessons {
			lessonDto, _ := NewLessonFromEntity(lesson)
			lessonDto.Normalize()
			lessonsDto = append(lessonsDto, lessonDto)
		}
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).
			Times(len(lessons))
		batchResults.On("Close").Once().Return(nil)

		// mock upsert teachers
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).
			Times(len(lessons)*len(recurringLesson.GetBaseLesson().GetTeacherIDs()) + len(lessons))
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).
			Times(len(lessons)*len(recurringLesson.GetBaseLesson().Classrooms.GetIDs()) + len(lessons))
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		for _, lesson := range lessons {
			mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
				Once().
				Return(batchResults, nil)
			for i := 0; i < len(lesson.Learners)+1; i++ {
				batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
			}
			batchResults.On("Close").Once().Return(nil)
		}

		// mock upsert lesson course
		for i := 0; i < len(lessons); i++ {
			mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
				Once().
				Return(batchResults, nil)
			for i := 0; i < len(recurringLesson.GetLessonCourses())+1; i++ {
				batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
			}
			batchResults.On("Close").Once().Return(nil)
		}
		lesson, err := r.UpsertLessons(ctx, mockDB.DB, recurringLesson)
		require.NoError(t, err)
		require.NotEmpty(t, lesson)
	})

	t.Run("error", func(t *testing.T) {
		baseLesson := &domain.Lesson{
			LessonID:         idutil.ULIDNow(),
			Name:             "lesson name",
			LocationID:       "center id",
			CreatedAt:        now,
			UpdatedAt:        now,
			StartTime:        now,
			EndTime:          now,
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
			TeachingMedium:   domain.LessonTeachingMediumOffline,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusAttend,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id-1",
				},
				{
					TeacherID: "teacher-id-2",
				},
			},
			Material: &domain.LessonMaterial{
				MediaIDs: []string{"media-id-1", "media-id-2"},
			},
			Classrooms: domain.LessonClassrooms{
				{
					ClassroomID: "classroom-id-1",
				},
				{
					ClassroomID: "classroom-id-2",
				},
			},
		}
		lessons := []*domain.Lesson{baseLesson}
		for i := 1; i < 5; i++ {
			lesson := *baseLesson
			lesson.LessonID = idutil.ULIDNow()
			lesson.StartTime = lesson.StartTime.AddDate(0, 0, 7)
			lesson.EndTime = lesson.StartTime.AddDate(0, 0, 7)
			learners := domain.LessonLearners{}
			for _, learner := range lesson.Learners {
				l := *learner
				l.AttendStatus = domain.StudentAttendStatusEmpty
				learners = append(learners, &l)
			}
			lesson.Learners = learners
			lesson.Material = &domain.LessonMaterial{}
			lessons = append(lessons, &lesson)
		}
		recurringLesson := &domain.RecurringLesson{
			Lessons: lessons,
		}
		for _, lesson := range lessons {
			// mock insert lesson gr
			gr := &LessonGroup{
				CourseID: database.Text("course-id-1"),
				MediaIDs: pgtype.TextArray{Status: pgtype.Null},
			}
			if len(lesson.Material.MediaIDs) != 0 {
				gr.MediaIDs = database.TextArray([]string{"media-id-1", "media-id-2"})
			}
			args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")},
				mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
			mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		}
		// mock insert lesson
		lessonsDto := make([]*Lesson, 0)
		for _, lesson := range recurringLesson.Lessons {
			lessonDto, _ := NewLessonFromEntity(lesson)
			lessonDto.Normalize()
			lessonsDto = append(lessonsDto, lessonDto)
		}
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).
			Times(len(lessons))
		batchResults.On("Close").Once().Return(nil)

		// mock upsert teachers
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).
			Times(len(lessons)*len(recurringLesson.GetBaseLesson().GetTeacherIDs()) + len(lessons))
		batchResults.On("Close").Once().Return(nil)

		// mock upsert classrooms
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).
			Times(len(lessons)*len(recurringLesson.GetBaseLesson().Classrooms.GetIDs()) + len(lessons))
		batchResults.On("Close").Once().Return(nil)

		// mock upsert members
		for _, lesson := range lessons {
			mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
				Once().
				Return(batchResults, nil)
			for i := 0; i < len(lesson.Learners)+1; i++ {
				batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Once()
			}
			batchResults.On("Close").Once().Return(nil)
		}
		// mock upsert lesson course
		for i := 0; i < len(lessons); i++ {
			mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
				Once().
				Return(batchResults, nil)
			for i := 0; i < len(recurringLesson.GetLessonCourses())+1; i++ {
				batchResults.On("Exec").Return(pgconn.CommandTag("1"), errors.New("internal error")).Once()
			}
			batchResults.On("Close").Once().Return(nil)
		}
		lesson, err := r.UpsertLessons(ctx, mockDB.DB, recurringLesson)
		require.Error(t, err)
		require.Empty(t, lesson)
	})
}

func TestLessonRepo_Retrieves(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	now := time.Now().UTC()
	args := &payloads.GetLessonListArg{
		CurrentTime: now,
		SchoolID:    "5",
		Compare:     ">=",
		LessonTime:  "future",
		Limit:       2,
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, _, _, _, err := r.Retrieve(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &Lesson{}
		selectFields := []string{"lesson_id", "name", "start_time", "end_time", "teaching_method", "teaching_medium", "center_id", "course_id", "class_id", "scheduling_status", "lesson_capacity", "end_at", "zoom_link", "classdo_link"}
		value := append(database.GetScanFields(e, selectFields))
		selectFields = append(selectFields)
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessons, _, _, _, err := r.Retrieve(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.EqualValues(t, []*domain.Lesson{
			{LessonID: "id", Persisted: true},
		}, lessons)

		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestLessonRepo_GetLessonBySchedulerID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schedulerID := "scheduler-id"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, schedulerID)
	repo, mockDB := LessonRepoWithSqlMock()
	lesson := &Lesson{}
	lessonFields, lessonValues := lesson.FieldMap()
	lessonMemberFields, lessonMemberValues := (&entities.LessonMember{}).FieldMap()
	e := &LessonGroup{}
	fields, values := e.FieldMap()
	lessonTeacherFields, lessonTeacherValues := (&LessonTeacher{}).FieldMap()
	lessonClassroomFields, lessonClassroomValues := (&LessonClassroom{}).FieldMap()
	reallocationFields, reallocationValues := (&Reallocation{}).FieldMap()

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanFields(nil, lessonFields, lessonValues)

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, lessonTeacherFields, [][]interface{}{
			lessonTeacherValues,
		})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, lessonMemberFields, [][]interface{}{
			lessonMemberValues,
		})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, lessonClassroomFields, [][]interface{}{
			lessonClassroomValues,
		})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, reallocationFields, [][]interface{}{
			reallocationValues,
		})

		lessonChain, err := repo.GetLessonBySchedulerID(ctx, mockDB.DB, schedulerID)
		assert.NoError(t, err)
		assert.NotEmpty(t, lessonChain)
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, args...)
		lessonChain, err := repo.GetLessonBySchedulerID(ctx, mockDB.DB, schedulerID)
		assert.Error(t, err)
		assert.Empty(t, lessonChain)
	})
}

func TestLessonRepo_GetLessonIdsByClassIdWithStartAndEndDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	now := time.Now().UTC()
	args := &domain.QueryLesson{
		ClassID:   "1",
		StartTime: &now,
	}

	t.Run("err select", func(t *testing.T) {

		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, err := r.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		e := &Lesson{}
		selectFields, _ := e.FieldMap()
		value := append(database.GetScanFields(e, selectFields))
		selectFields = append(selectFields)
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})
		lessons, err := r.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.EqualValues(t, []*domain.Lesson{
			{LessonID: "id", Persisted: true},
		}, lessons)
	})
}

func TestLessonRepo_UpdateSchedulingStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LessonRepoWithSqlMock()
	lessonStatus := map[string]domain.LessonSchedulingStatus{
		"lesson-id-1": domain.LessonSchedulingStatusPublished,
		"lesson-id-2": domain.LessonSchedulingStatusPublished,
	}
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), errors.New("internal error")).Times(len(lessonStatus))
		batchResults.On("Close").Once().Return(nil)
		err := r.UpdateSchedulingStatus(ctx, mockDB.DB, lessonStatus)
		assert.Error(t, err, "internal error")
	})
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Times(len(lessonStatus))
		batchResults.On("Close").Once().Return(nil)
		err := r.UpdateSchedulingStatus(ctx, mockDB.DB, lessonStatus)
		assert.Nil(t, err)
	})
}

func TestLessonRepo_GetLessonsOnCalendar(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockLessonRepo, mockDB := LessonRepoWithSqlMock()

	now := time.Now().UTC()
	args := &payloads.GetLessonListOnCalendarArgs{
		View:                                payloads.Weekly,
		FromDate:                            now,
		ToDate:                              now.Add(7 * 24 * time.Hour),
		LocationID:                          "test-location-1",
		Timezone:                            "sample-timezone",
		IsIncludeNoneAssignedTeacherLessons: true,
	}

	fields := []string{
		"lesson_id",
		"name",
		"start_time",
		"end_time",
		"teaching_method",
		"teaching_medium",
		"center_id",
		"course_id",
		"class_id",
		"scheduling_status",
		"scheduler_id",
		"lesson_capacity",
		"name",
		"name",
	}
	lesson := &Lesson{}
	var (
		courseName pgtype.Text
		className  pgtype.Text
	)
	values := append(database.GetScanFields(lesson, fields), &courseName, &className)

	t.Run("failed to get lessons on calendar", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)

		lessons, err := mockLessonRepo.GetLessonsOnCalendar(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("successfully fetched lessons on calendar with weekly/monthly view", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessons, err := mockLessonRepo.GetLessonsOnCalendar(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.NotNil(t, lessons)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})

	t.Run("successfully fetched lessons on calendar with daily view", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		args.View = payloads.Daily
		lessons, err := mockLessonRepo.GetLessonsOnCalendar(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.NotNil(t, lessons)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonRepo_GetLessonWithNamesByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockLessonRepo, mockDB := LessonRepoWithSqlMock()
	lessonID := "lesson-id1"

	fields := []string{
		"lesson_id",
		"name",
		"start_time",
		"end_time",
		"teaching_method",
		"teaching_medium",
		"center_id",
		"course_id",
		"class_id",
		"scheduling_status",
		"scheduler_id",
		"is_locked",
		"zoom_id",
		"zoom_link",
		"zoom_owner_id",
		"classdo_owner_id",
		"classdo_link",
		"classdo_room_id",
		"lesson_capacity",
		"name",
		"name",
		"name",
	}
	lesson := &Lesson{}
	var (
		courseName   pgtype.Text
		className    pgtype.Text
		locationName pgtype.Text
	)
	values := append(database.GetScanFields(lesson, fields), &courseName, &className, &locationName)

	t.Run("failed get lesson", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, fields, values)

		lesson, err := mockLessonRepo.GetLessonWithNamesByID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lesson)
	})

	t.Run("successfully get lesson", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, values)

		lesson, err := mockLessonRepo.GetLessonWithNamesByID(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
		assert.NotNil(t, lesson)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonRepo_GetLessonsByLocationStatusAndDateTimeRange(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	mockLessonRepo, mockDB := LessonRepoWithSqlMock()
	params := &payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
		LocationID:   "location-id1",
		LessonStatus: domain.LessonSchedulingStatusCompleted,
		StartDate:    now,
		EndDate:      now.Add(12 * 24 * time.Hour),
		StartTime:    now,
		EndTime:      now.Add(4 * time.Hour),
		Timezone:     "timezone",
	}

	lesson := &Lesson{}
	fields, values := lesson.FieldMap()

	t.Run("failed get lessons", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, err := mockLessonRepo.GetLessonsByLocationStatusAndDateTimeRange(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("successfully get lessons", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)

		lessons, err := mockLessonRepo.GetLessonsByLocationStatusAndDateTimeRange(ctx, mockDB.DB, params)
		assert.Nil(t, err)
		assert.NotNil(t, lessons)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonRepo_RemoveZoomLinkByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LessonRepoWithSqlMock()
	lessonID := "lessonID"
	t.Run("error", func(t *testing.T) {

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &lessonID, mock.AnythingOfType("Time"))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.RemoveZoomLinkByLessonID(ctx, mockDB.DB, lessonID)
		assert.Error(t, err, "internal error")
	})
	t.Run("success", func(t *testing.T) {

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &lessonID, mock.AnythingOfType("Time"))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.RemoveZoomLinkByLessonID(ctx, mockDB.DB, "lessonID")
		assert.Nil(t, err)
	})
}

func TestLessonRepo_GetFutureLessonsByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lesson := &Lesson{}

	tz := "UTC"
	courseIDs := []string{"course-id-1", "course-id-2"}
	lessonFields, lessonValues := lesson.FieldMap()
	lessonTeacher := &LessonTeacher{}
	lessonTeacherFields, lessonTeacherValues := lessonTeacher.FieldMap()
	lessonMemberFields, lessonMemberValues := (&entities.LessonMember{}).FieldMap()

	t.Run("err select", func(t *testing.T) {
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, err := r.GetFutureLessonsByCourseIDs(ctx, mockDB.DB, courseIDs, tz)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success", func(t *testing.T) {
		//lesson
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, courseIDs, tz)
		mockDB.MockScanFields(nil, lessonFields, lessonValues)

		//lesson-teacher
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, lessonTeacherFields, lessonTeacherValues)

		// //lesson-learner
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, lessonMemberFields, [][]interface{}{
			lessonMemberValues,
		})

		gotLessons, err := r.GetFutureLessonsByCourseIDs(ctx, mockDB.DB, courseIDs, tz)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessons)
	})
}

func TestLessonTeacherRepo_UpdateLessonTeacherNameq(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lessons := []*domain.Lesson{
		{
			LessonID:        "lesson-id-1",
			CourseID:        "course-id-1",
			PreparationTime: 120,
			BreakTime:       15,
		},
		{
			LessonID:        "lesson-id-2",
			CourseID:        "course-id-2",
			PreparationTime: 150,
			BreakTime:       10,
		},
	}
	gr := &LessonGroup{
		CourseID: database.Text("course-id-1"),
		MediaIDs: database.TextArray([]string{"media-id-1", "media-id-2"}),
	}

	t.Run("err update", func(t *testing.T) {
		// mock insert lesson gr
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)

		err := r.UpdateLessonsTeachingTime(ctx, mockDB.DB, lessons)
		assert.Equal(t, "failed to upsert lesson: batchResults.Exec: batchResults.Exec: closed pool", err.Error())
	})

	t.Run("success", func(t *testing.T) {
		// mock insert lesson gr
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, &gr.CourseID, &gr.MediaIDs, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// mock upsert lessons
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Twice().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.UpdateLessonsTeachingTime(ctx, mockDB.DB, lessons)
		assert.Equal(t, nil, err)
	})
}

func TestLessonRepo_GetLessonsWithSchedulerNull(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lesson := &Lesson{}
	limit := 10
	offset := 0

	lessonFields, lessonValues := lesson.FieldMap()

	t.Run("err select", func(t *testing.T) {
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, err := r.GetLessonsWithSchedulerNull(ctx, mockDB.DB, 10, 0)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success", func(t *testing.T) {
		//lesson
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &limit, &offset)
		mockDB.MockScanFields(nil, lessonFields, lessonValues)

		gotLessons, err := r.GetLessonsWithSchedulerNull(ctx, mockDB.DB, limit, offset)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessons)
	})
}

func TestLessonRepo_FillSchedulerToLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LessonRepoWithSqlMock()
	schedulerMaps := map[string]string{
		"lesson_id_01": "scheduler_id_01",
		"lesson_id_02": "scheduler_id_02",
	}
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), errors.New("internal error")).Times(len(schedulerMaps))
		batchResults.On("Close").Once().Return(nil)
		err := r.FillSchedulerToLessons(ctx, mockDB.DB, schedulerMaps)
		assert.Error(t, err, "internal error")
	})
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil).Times(len(schedulerMaps))
		batchResults.On("Close").Once().Return(nil)
		err := r.FillSchedulerToLessons(ctx, mockDB.DB, schedulerMaps)
		assert.Nil(t, err)
	})
}

func TestLessonRepo_GetLessonsWithInvalidSchedulerID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lesson := &Lesson{}

	lessonFields, lessonValues := lesson.FieldMap()

	t.Run("err select", func(t *testing.T) {
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, err := r.GetLessonsWithInvalidSchedulerID(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success", func(t *testing.T) {
		//lesson
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, lessonFields, lessonValues)

		gotLessons, err := r.GetLessonsWithInvalidSchedulerID(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessons)
	})
}
