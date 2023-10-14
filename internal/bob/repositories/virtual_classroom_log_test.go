package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func VirtualClassroomLogRepoWithSqlMock() (*VirtualClassroomLogRepo, *testutil.MockDB) {
	r := &VirtualClassroomLogRepo{}
	return r, testutil.NewMockDB()
}

func TestVirtualClassroomLogRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualClassroomLogRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities.VirtualClassRoomLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities.VirtualClassRoomLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, err, "cannot insert new VirtualClassroomLog")
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.VirtualClassRoomLog{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestVirtualClassroomLogRepo_AddAttendeeIDByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualClassroomLogRepoWithSqlMock()

	t.Run("err update", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		attendeeID := database.Text("user-id-1")
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID, attendeeID)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.AddAttendeeIDByLessonID(ctx, mockDB.DB, lessonID, attendeeID)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		attendeeID := database.Text("user-id-1")
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID, attendeeID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.AddAttendeeIDByLessonID(ctx, mockDB.DB, lessonID, attendeeID)
		assert.Nil(t, err)
	})
}

func TestVirtualClassroomLogRepo_IncreaseTotalTimesByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualClassroomLogRepoWithSqlMock()

	t.Run("err update", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		logType := entities.TotalTimesReconnection
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.IncreaseTotalTimesByLessonID(ctx, mockDB.DB, lessonID, logType)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		logType := entities.TotalTimesGettingRoomState
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.IncreaseTotalTimesByLessonID(ctx, mockDB.DB, lessonID, logType)
		assert.Nil(t, err)
	})

	t.Run("wrong log type", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		logType := entities.TotalTimes(10000)
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.IncreaseTotalTimesByLessonID(ctx, mockDB.DB, lessonID, logType)
		assert.EqualError(t, err, fmt.Sprintf("not handle this type yet %v", logType))
	})
}

func TestVirtualClassroomLogRepo_CompleteLogByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualClassroomLogRepoWithSqlMock()

	t.Run("err update", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.CompleteLogByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		lessonID := database.Text("lesson-id-1")
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.CompleteLogByLessonID(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
	})
}

func TestVirtualClassroomLogRepo_GetLatestByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualClassroomLogRepoWithSqlMock()

	ID := idutil.ULIDNow()
	lessonID := database.Text(ID)

	t.Run("no row err", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&lessonID,
		)

		e := &entities.VirtualClassRoomLog{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		results, err := r.GetLatestByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, results)
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&lessonID,
		)

		e := &entities.VirtualClassRoomLog{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)

		results, err := r.GetLatestByLessonID(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
		assert.Equal(t, e, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
