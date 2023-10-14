package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonMemberRepoWithSqlMock() (*LessonMemberRepo, *testutil.MockDB) {
	r := &LessonMemberRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonMemberRepo_ListStudentsByLessonArgs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	args := &domain.ListStudentsByLessonArgs{
		LessonID: "lesson-1",
		Limit:    10,
		UserName: "username",
		UserID:   "userid",
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &args.LessonID, &args.UserName, &args.UserID, &args.Limit)

		lessonMembers, err := l.ListStudentsByLessonArgs(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonMembers)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &args.LessonID, &args.UserName, &args.UserID, &args.Limit)

		e := &User{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := l.ListStudentsByLessonArgs(ctx, mockDB.DB, args)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonMemberRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	lessonID, userID := "lessonId", "studentID"

	t.Run("err", func(t *testing.T) {
		e := &LessonMember{}
		fields, values := e.FieldMap()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, lessonID, userID)
		mockDB.MockRowScanFields(errors.New("error"), fields, values)

		lessonMembers, err := l.FindByID(ctx, mockDB.DB, lessonID, userID)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
		assert.Nil(t, lessonMembers)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, lessonID, userID)

		e := &LessonMember{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)

		_, err := l.FindByID(ctx, mockDB.DB, lessonID, userID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonMemberRepo_GetLessonIDsByStudentCourseRemovedLocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	courseID := "course_id_1"
	studentID := "student_id_1"
	locationIDs := []string{"location-1", "location-2"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, mock.Anything, mock.Anything, &courseID, &studentID, &locationIDs)

		lessonIDs, err := l.GetLessonIDsByStudentCourseRemovedLocation(ctx, mockDB.DB, courseID, studentID, locationIDs)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.Nil(t, lessonIDs)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&courseID, &studentID, &locationIDs,
		)

		e := &LessonMember{}
		fields, _ := e.FieldMap()
		values := []interface{}{&e.UserID.String}
		mockDB.MockScanArray(nil, []string{fields[3]}, [][]interface{}{values})

		lessonIDs, err := l.GetLessonIDsByStudentCourseRemovedLocation(ctx, mockDB.DB, courseID, studentID, locationIDs)

		assert.Nil(t, err)
		assert.NotNil(t, lessonIDs)
	})
}

func TestLessonMemberRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonMemberRepoWithSqlMock()

	studentID := "studentID"
	lessonIDs := []string{"lesson-1", "lesson-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID, &lessonIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.SoftDelete(ctx, mockDB.DB, studentID, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID, &lessonIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.SoftDelete(ctx, mockDB.DB, studentID, lessonIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "lesson_members")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"lesson_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestLessonMemberRepo_GetLessonMembersInLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, ids)

		lessonMembers, err := l.GetLessonMembersInLessons(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonMembers)
	})

	t.Run("success with selec", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, ids)

		e := &LessonMember{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := l.GetLessonMembersInLessons(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"lesson_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLessonMemberRepo_InsertLessonMembers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	lessons := []*domain.LessonMember{{LessonID: "1", StudentID: ""}}

	t.Run("err select", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := l.InsertLessonMembers(ctx, mockDB.DB, lessons)
		assert.Equal(t, "batchResults.Exec: batchResults.Exec: closed pool", err.Error())

	})

	t.Run("happy case", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := l.InsertLessonMembers(ctx, mockDB.DB, lessons)
		assert.Equal(t, nil, err)

	})
}

func TestLessonMemberRepo_DeleteLessonMembers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	lessons := []*domain.LessonMember{{LessonID: "1", StudentID: ""}}

	t.Run("err select", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := l.DeleteLessonMembers(ctx, mockDB.DB, lessons)
		assert.Equal(t, "batchResults.Exec: batchResults.Exec: closed pool", err.Error())

	})

	t.Run("happy case", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := l.DeleteLessonMembers(ctx, mockDB.DB, lessons)
		assert.Equal(t, nil, err)

	})
}

func TestLessonMemberRepo_UpdateLessonMembers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	lessonReports := []*domain.UpdateLessonMemberReport{{LessonID: "1", StudentID: "1",
		AttendanceStatus: "", AttendanceNotice: "", AttendanceReason: "", AttendanceNote: ""}}

	t.Run("err select", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := l.UpdateLessonMembers(ctx, mockDB.DB, lessonReports)
		assert.Equal(t, "batchResults.Exec: batchResults.Exec: closed pool", err.Error())

	})

	t.Run("happy case", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := l.UpdateLessonMembers(ctx, mockDB.DB, lessonReports)
		assert.Equal(t, nil, err)

	})
}

func TestLessonMemberRepo_GetLessonsOutOfStudentCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	l, mockDB := LessonMemberRepoWithSqlMock()
	startAt := time.Date(2022, 8, 29, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2022, 8, 30, 9, 0, 0, 0, time.UTC)
	studentCourse := &user_domain.StudentSubscription{
		StudentID: "student-id",
		StartAt:   startAt,
		EndAt:     endAt,
		CourseID:  "course-id",
	}
	args := []interface{}{mock.Anything, mock.Anything, "course-id", "student-id", startAt, endAt}
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, args...)
		lessonMembers, err := l.GetLessonsOutOfStudentCourse(ctx, mockDB.DB, studentCourse)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonMembers)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		var lessonID string
		mockDB.MockScanArray(nil, []string{"lesson_id"}, [][]interface{}{
			{&lessonID},
		})
		lessonIDs, err := l.GetLessonsOutOfStudentCourse(ctx, mockDB.DB, studentCourse)
		fmt.Println(lessonID)
		assert.Nil(t, err)
		assert.NotNil(t, lessonIDs)
	})
}

func TestLessonMemberRepo_DeleteLessonMembersByStartDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonMemberRepoWithSqlMock()

	studentID := "studentID"
	classID := "class-1"
	endTime := time.Now()

	t.Run("err delete", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID, &classID, &endTime)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, args...)

		_, err := r.DeleteLessonMembersByStartDate(ctx, mockDB.DB, studentID, classID, endTime)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID, &classID, &endTime)
		mockDB.MockQueryArgs(t, nil, args...)
		var lessonID pgtype.Text
		mockDB.MockScanArray(nil, []string{"lesson_id"}, [][]interface{}{
			{&lessonID},
		})
		_, err := r.DeleteLessonMembersByStartDate(ctx, mockDB.DB, studentID, classID, endTime)
		assert.Nil(t, err)
	})
}

func TestLessonMemberRepo_UpdateLessonMemberName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	lessonMembers := []*domain.UpdateLessonMemberName{{LessonID: "lesson-1", StudentID: "user-1",
		UserFirstName: "First name", UserLastName: "Last name"}}

	t.Run("err update", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := l.UpdateLessonMemberNames(ctx, mockDB.DB, lessonMembers)
		assert.Equal(t, "batchResults.Exec: batchResults.Exec: closed pool", err.Error())
	})

	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := l.UpdateLessonMemberNames(ctx, mockDB.DB, lessonMembers)
		assert.Equal(t, nil, err)
	})
}

func TestLessonMemberRepo_GetLessonLearnersWithCourseAndNamesByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockLessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()
	lessonIDs := []string{"test-lesson-id-1", "test-lesson-id-2", "test-lesson-id-3"}
	lessonMember := &LessonMember{}
	var courseName, name pgtype.Text
	fields := []string{
		"lesson_id",
		"user_id",
		"course_id",
		"attendance_status",
		"attendance_notice",
		"attendance_reason",
		"attendance_note",
		"course_name",
		"name",
	}
	values := append(database.GetScanFields(lessonMember, fields), &courseName, &name)

	t.Run("failed to get lesson learners", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		lessonLearners, err := mockLessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx, mockDB.DB, lessonIDs, false)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonLearners)
	})

	t.Run("successfully fetched lesson learners", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonLearners, err := mockLessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx, mockDB.DB, lessonIDs, false)
		assert.Nil(t, err)
		assert.NotNil(t, lessonLearners)
	})

	t.Run("successfully fetched lesson learners using user public info", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonLearners, err := mockLessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx, mockDB.DB, lessonIDs, true)
		assert.Nil(t, err)
		assert.NotNil(t, lessonLearners)
	})
}
