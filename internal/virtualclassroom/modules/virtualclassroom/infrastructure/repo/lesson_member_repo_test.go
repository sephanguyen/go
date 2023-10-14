package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vl_payloads "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	args := &MemberStatesFilter{
		LessonID:  database.Text("lesson-1"),
		UserID:    database.Text("user-1"),
		StateType: database.Text("type"),
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &args.LessonID, &args.UserID, &args.StateType)

		lessonMembers, err := l.GetLessonMemberStatesWithParams(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonMembers)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &args.LessonID, &args.UserID, &args.StateType)

		e := &LessonMemberStateDTO{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := l.GetLessonMemberStatesWithParams(ctx, mockDB.DB, args)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonMemberRepo_UpsertLessonMemberState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &LessonGroupDTO{}
	_, fieldMap := mockE.FieldMap()

	r, mockDB := LessonGroupRepoWithSqlMock()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Upsert(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("err: upsert failed", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := r.Upsert(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("%w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	// t.Run("err: no row affected", func(t *testing.T) {
	// 	cmdTag := pgconn.CommandTag([]byte(`0`))
	// 	mockDB.MockExecArgs(t, cmdTag, nil, args...)

	// 	err := r.Upsert(ctx, mockDB.DB, mockE)
	// 	assert.NotNil(t, err)

	// 	mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	// })
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestLessonMemberRepo_GetCourseAccessible(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()
	var courseID pgtype.Text
	fields := []string{"course_id"}
	values := []interface{}{&courseID}
	userID := "user-id1"

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &userID)
		mockDB.MockScanFields(nil, fields, values)

		courseIDs, err := lessonMemberRepo.GetCourseAccessible(ctx, mockDB.DB, userID)
		assert.NoError(t, err)
		assert.NotNil(t, courseIDs)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &userID)

		courseIDs, err := lessonMemberRepo.GetCourseAccessible(ctx, mockDB.DB, userID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})
}

func TestLessonMemberRepo_GetLessonMemberStatesByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonMemberRepoWithSqlMock()
	lessonID := "lesson-1"

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything)

		lessonMemberStates, err := l.GetLessonMemberStatesByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonMemberStates)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)

		e := &LessonMemberStateDTO{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonMemberStates, err := l.GetLessonMemberStatesByLessonID(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
		assert.NotNil(t, lessonMemberStates)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLessonMemberRepo_GetLearnerIDsByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()

	lessonID := "lessonID"

	t.Run("successful", func(t *testing.T) {
		result := pgtype.Text{String: "user-id", Status: pgtype.Present}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonID)
		mockDB.MockScanFields(nil, []string{""}, []interface{}{&result})

		learners, err := lessonMemberRepo.GetLearnerIDsByLessonID(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, learners)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonID)

		learners, err := lessonMemberRepo.GetLearnerIDsByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, learners)
	})
}

func TestLessonMemberRepo_GetLearnersByLessonIDWithPaging(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()
	user := &LessonMemberDTO{}
	fields, values := user.FieldMap()

	t.Run("successful", func(t *testing.T) {
		params := &vl_payloads.GetLearnersByLessonIDArgs{
			LessonID:       "lesson-id1",
			Limit:          15,
			LessonCourseID: "lesson-id1course-id1",
			UserID:         "user-id1",
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.LessonID, params.LessonCourseID, params.UserID, params.Limit)
		mockDB.MockScanFields(nil, fields, values)

		learners, err := lessonMemberRepo.GetLearnersByLessonIDWithPaging(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, learners)
	})

	t.Run("failed", func(t *testing.T) {
		params := &vl_payloads.GetLearnersByLessonIDArgs{
			LessonID:       "lesson-id1",
			Limit:          15,
			LessonCourseID: "lesson-id1course-id1",
			UserID:         "user-id1",
		}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, params.LessonID, params.LessonCourseID, params.UserID, params.Limit)

		learners, err := lessonMemberRepo.GetLearnersByLessonIDWithPaging(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, learners)
	})
}

func TestLessonMemberRepo_GetLessonMemberUsersByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()
	user := &User{}
	fields, values := user.FieldMap()
	lessonID := "lesson-id-1"

	t.Run("successful with bob db", func(t *testing.T) {
		params := &vl_payloads.GetLessonMemberUsersByLessonIDArgs{
			LessonID:        lessonID,
			UseLessonmgmtDB: false,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonID)
		mockDB.MockScanFields(nil, fields, values)

		learners, err := lessonMemberRepo.GetLessonMemberUsersByLessonID(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, learners)
	})

	t.Run("successful with lessonmgmt db", func(t *testing.T) {
		params := &vl_payloads.GetLessonMemberUsersByLessonIDArgs{
			LessonID:        lessonID,
			StudentIDs:      []string{"studentID"},
			UseLessonmgmtDB: true,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.StudentIDs)
		mockDB.MockScanFields(nil, fields, values)

		learners, err := lessonMemberRepo.GetLessonMemberUsersByLessonID(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, learners)
	})

	t.Run("failed with bob db", func(t *testing.T) {
		params := &vl_payloads.GetLessonMemberUsersByLessonIDArgs{
			LessonID:        lessonID,
			UseLessonmgmtDB: false,
		}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonID)

		learners, err := lessonMemberRepo.GetLessonMemberUsersByLessonID(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, learners)
	})

	t.Run("failed with lessonmgmt db", func(t *testing.T) {
		params := &vl_payloads.GetLessonMemberUsersByLessonIDArgs{
			LessonID:        lessonID,
			StudentIDs:      []string{"studentID"},
			UseLessonmgmtDB: true,
		}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, params.StudentIDs)

		learners, err := lessonMemberRepo.GetLessonMemberUsersByLessonID(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, learners)
	})
}

func TestLessonMemberRepo_InsertMissingLessonMemberStateByState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()
	state := &StateValueDTO{}
	_, values := state.FieldMap()

	lessonID := "lesson-id1"
	stateType := domain.LearnerStateTypeChat
	state.BoolValue = database.Bool(true)
	state.StringArrayValue = database.TextArray([]string{})

	args := append([]interface{}{mock.Anything, mock.Anything, lessonID, domain.LearnerStateTypeChat}, values...)

	t.Run("failed", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := lessonMemberRepo.InsertMissingLessonMemberStateByState(ctx, mockDB.DB, lessonID, stateType, state)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("successful", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := lessonMemberRepo.InsertMissingLessonMemberStateByState(ctx, mockDB.DB, lessonID, stateType, state)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

}

func TestLessonMemberRepo_InsertLessonMemberState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo, mockDB := LessonMemberRepoWithSqlMock()

	now := time.Now()
	state := &LessonMemberStateDTO{
		LessonID:  database.Text("lesson-id1"),
		UserID:    database.Text("user-id1"),
		StateType: database.Text(string(domain.LearnerStateTypeChat)),
		CreatedAt: database.Timestamptz(now),
		UpdatedAt: database.Timestamptz(now),
		BoolValue: database.Bool(true),
	}
	_, values := state.FieldMap()

	args := append([]interface{}{mock.Anything, mock.Anything}, values...)

	t.Run("failed", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := lessonMemberRepo.InsertLessonMemberState(ctx, mockDB.DB, state)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("successful", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := lessonMemberRepo.InsertLessonMemberState(ctx, mockDB.DB, state)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

}

func TestLessonMemberRepo_GetLessonLearnersByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockRepo, mockDB := LessonMemberRepoWithSqlMock()
	lessonIDs := []string{"test-lesson-id-1", "test-lesson-id-2", "test-lesson-id-3"}
	dto := &LessonMemberDTO{}
	fields, values := dto.FieldMap()

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		lessonLearners, err := mockRepo.GetLessonLearnersByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonLearners)
	})

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonLearners, err := mockRepo.GetLessonLearnersByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.Nil(t, err)
		assert.NotNil(t, lessonLearners)
	})
}
