package repo

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ClassMemberRepoWithSqlMock() (*ClassMemberRepo, *testutil.MockDB) {
	classMemberRepo := &ClassMemberRepo{}
	return classMemberRepo, testutil.NewMockDB()
}

func TestClassMemberRepo_GetByClassIDAndUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classMemberRepo, mockDB := ClassMemberRepoWithSqlMock()
	e := &ClassMember{}
	fields, value := e.FieldMap()
	classID := idutil.ULIDNow()
	userID := "user-1"
	t.Run("error no row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, value)
		gotClass, err := classMemberRepo.GetByClassIDAndUserID(ctx, mockDB.DB, classID, userID)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, gotClass)
	})
	t.Run("error tx closed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, value)
		gotClass, err := classMemberRepo.GetByClassIDAndUserID(ctx, mockDB.DB, classID, userID)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
		assert.Nil(t, gotClass)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, value)
		gotClass, err := classMemberRepo.GetByClassIDAndUserID(ctx, mockDB.DB, classID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, gotClass)
	})
}

func TestClassMemberRepo_GetByClassIDAndUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassMemberRepoWithSqlMock()
	classID := "user-1"
	userIDs := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &classID, &userIDs)

		locations, err := r.GetByClassIDAndUserIDs(ctx, mockDB.DB, classID, userIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &classID, &userIDs)

		e := &ClassMember{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByClassIDAndUserIDs(ctx, mockDB.DB, classID, userIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"class_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestClassMemberRepo_UpsertClassMembers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	classMemberRepo, mockDB := ClassMemberRepoWithSqlMock()
	classes := []*domain.ClassMember{
		{
			ClassID:       "class-1",
			ClassMemberID: "class-member-1",
			UserID:        "user-1",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ClassID:       "class-2",
			ClassMemberID: "class-member-2",
			UserID:        "user-2",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := classMemberRepo.UpsertClassMembers(ctx, mockDB.DB, classes)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := classMemberRepo.UpsertClassMembers(ctx, mockDB.DB, classes)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestClassMemberRepo_UpsertClassMember(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	classMemberRepo, mockDB := ClassMemberRepoWithSqlMock()
	classMember := &domain.ClassMember{
		ClassID:       "class-1",
		ClassMemberID: "class-member-1",
		UserID:        "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	dto, _ := NewClassMemberFromEntity(classMember)
	_, values := dto.FieldMap()
	t.Run("success", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, cmdTag, nil, args...)
		err := classMemberRepo.UpsertClassMember(ctx, mockDB.DB, classMember)
		require.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, cmdTag, puddle.ErrNotAvailable, args...)
		err := classMemberRepo.UpsertClassMember(ctx, mockDB.DB, classMember)
		require.ErrorIs(t, err, puddle.ErrNotAvailable)
	})
}

func TestClassRepo_DeleteByUserIDAndClassID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classMemberRepo, mockDB := ClassMemberRepoWithSqlMock()
	userID := "user-id-1"
	classID := "class-id-1"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, userID, classID)
	t.Run("error", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)
		err := classMemberRepo.DeleteByUserIDAndClassID(ctx, mockDB.DB, userID, classID)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := classMemberRepo.DeleteByUserIDAndClassID(ctx, mockDB.DB, userID, classID)
		assert.NoError(t, err)
	})
}

func TestClassMember_FindStudentIDWithCourseIDsByClassIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassMemberRepoWithSqlMock()

	classIds := []string{"class-id-1", "class-id-2"}
	studentIds := []string{"student-id-1", "student-id-2"}
	courseIds := []string{"course-id-1", "course-id-2"}
	expectedResult := []string{"student-id-1", "course-id-1", "student-id-2", "course-id-2"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &classIds)

		result, err := r.FindStudentIDWithCourseIDsByClassIDs(ctx, mockDB.DB, classIds)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, &classIds).Once().Return(rows, nil)
		rows.On("Next").Times(2).Return(true)
		y := 0
		for i := 0; i < 2; i++ {
			var studentID, courseID string
			rows.On("Scan", &studentID, &courseID).Once().Run(func(args mock.Arguments) {
				reflect.ValueOf(args[0]).Elem().SetString(studentIds[y])
				reflect.ValueOf(args[1]).Elem().SetString(courseIds[y])
				y++
			}).Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		result, err := r.FindStudentIDWithCourseIDsByClassIDs(ctx, mockDB.DB, classIds)
		assert.Nil(t, err)
		assert.EqualValues(t, result, expectedResult)

	})
}

func TestClassMemberRepo_GetByUserAndCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassMemberRepoWithSqlMock()
	userID := "user-1"
	courseID := "course-1"
	classID := "class-1"

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &userID, &courseID)

		classMember, err := r.GetByUserAndCourse(ctx, mockDB.DB, userID, courseID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, classMember)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &userID, &courseID)
		e := &ClassMember{
			ClassMemberID: database.Text("id-1"),
			UserID:        database.Text(userID),
			ClassID:       database.Text("class-1"),
		}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		classMember, err := r.GetByUserAndCourse(ctx, mockDB.DB, userID, courseID)
		assert.NoError(t, err)
		assert.NotNil(t, classMember)
		assert.Len(t, classMember, 1)
		for k, v := range classMember {
			assert.Equal(t, userID, k)
			assert.Equal(t, "id-1", v.ClassMemberID)
			assert.Equal(t, classID, v.ClassID)

		}
	})
}

func TestLessonMember_RetrieveByClassIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c, mockDB := ClassMemberRepoWithSqlMock()
	filter := &queries.FindClassMemberFilter{
		ClassIDs: []string{"1", "2"},
		Limit:    1,
		OffsetID: "3",
		UserName: "name",
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &filter.ClassIDs, &filter.UserName, &filter.OffsetID, &filter.Limit)

		result, err := c.RetrieveByClassIDs(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &filter.ClassIDs, &filter.UserName, &filter.OffsetID, &filter.Limit)

		e := &ClassMember{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := c.RetrieveByClassIDs(ctx, mockDB.DB, filter)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}
