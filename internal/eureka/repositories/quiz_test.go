package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func QuizRepoWithSqlMock() (*QuizRepo, *testutil.MockDB) {
	r := &QuizRepo{}
	return r, testutil.NewMockDB()
}

func TestQuizRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities.Quiz{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities.Quiz{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Equal(t, fmt.Errorf("can not create quiz"), err)
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.Quiz{}
		fields, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestQuizRepo_Search(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	filter := QuizFilter{
		ExternalIDs: database.TextArray([]string{"externalID"}),
		Status:      database.Text("status"),
		Limit:       uint(10),
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&filter.ExternalIDs,
			&filter.Status,
		)

		results, err := r.Search(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&filter.ExternalIDs,
			&filter.Status,
		)

		e := &entities.Quiz{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		results, err := r.Search(ctx, mockDB.DB, filter)
		assert.Nil(t, err)
		assert.Equal(t, entities.Quizzes{e}, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":  {HasNullTest: true},
			"external_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"status":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}

func TestQuizRepo_GetByExternalID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	externalID := idutil.ULIDNow()
	pgExternalID := database.Text(externalID)
	schoolID := database.Int4(1)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgExternalID,
			&schoolID,
		)

		results, err := r.GetByExternalID(ctx, mockDB.DB, pgExternalID, schoolID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgExternalID,
			&schoolID,
		)

		e := &entities.Quiz{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanFields(nil, fields, values)

		sPackage, err := r.GetByExternalID(ctx, mockDB.DB, pgExternalID, schoolID)
		assert.Nil(t, err)
		assert.Equal(t, e, sPackage)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

// ShuffledQuizSetRepoWithSQLMock test repo with mock
func ShuffledQuizSetRepoWithSQLMock() (*ShuffledQuizSetRepo, *testutil.MockDB) {
	r := &ShuffledQuizSetRepo{}
	return r, testutil.NewMockDB()
}

func TestQuizRepo_GetByExternalIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	pgExternalIDs := database.TextArray([]string{"externalIDs"})
	loID := database.Text(idutil.ULIDNow())

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgExternalIDs,
			&loID,
		)

		results, err := r.GetByExternalIDsAndLOID(ctx, mockDB.DB, pgExternalIDs, loID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgExternalIDs,
			&loID,
		)

		e := &entities.Quiz{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		sPackage, err := r.GetByExternalIDsAndLOID(ctx, mockDB.DB, pgExternalIDs, loID)
		assert.Nil(t, err)
		assert.Equal(t, entities.Quizzes{e}, sPackage)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestQuizRepo_GetByExternalIDsAndLmID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	externalIDs := database.TextArray([]string{"external-id"})
	lmID := database.Text("learning-material-id")
	quizID := database.Text(idutil.ULIDNow())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.AnythingOfType("string"),
					&externalIDs,
					&lmID,
				)

				e := &entities.Quiz{
					ID: quizID,
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			expectedResp: entities.Quizzes{
				&entities.Quiz{
					ID: quizID,
				},
			},
		},
		{
			name: "select error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
					mock.AnythingOfType("string"),
					&externalIDs,
					&lmID,
				)
			},
			expectedErr: fmt.Errorf("database.Select: err db.Query: %w", puddle.ErrClosedPool),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := r.GetByExternalIDsAndLmID(ctx, mockDB.DB, externalIDs, lmID)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(entities.Quizzes), resp)
			}
		})
	}
}

func TestQuizRepo_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	quizID := idutil.ULIDNow()
	pgQuizID := database.Text(quizID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgQuizID,
		)

		results, err := r.Retrieve(ctx, mockDB.DB, pgQuizID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgQuizID,
		)

		e := &entities.Quiz{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanFields(nil, fields, values)

		sPackage, err := r.Retrieve(ctx, mockDB.DB, pgQuizID)
		assert.Nil(t, err)
		assert.Equal(t, e, sPackage)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestQuizRepo_GetOptions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	quizID := database.Text(idutil.ULIDNow())
	loID := database.Text("lo_id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			quizID,
			loID,
		)

		quiz := &entities.Quiz{}
		fields := []string{"options"}
		values := []interface{}{&quiz.Options}
		mockDB.MockRowScanFields(puddle.ErrClosedPool, fields, values)

		results, err := r.GetOptions(ctx, mockDB.DB, quizID, loID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			quizID,
			loID,
		)

		quiz := &entities.Quiz{}
		fields := []string{"options"}
		values := []interface{}{&quiz.Options}

		options := []*entities.QuizOption{}
		quiz.Options.AssignTo(&options)
		mockDB.MockRowScanFields(nil, fields, values)
		results, err := r.GetOptions(ctx, mockDB.DB, quizID, loID)
		assert.Nil(t, err)
		assert.Equal(t, options, results)
	})
}

func TestQuizRepo_DeleteByExternalID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	id := database.Text("externalID")
	schoolID := database.Int4(1)
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &id, &schoolID)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.DeleteByExternalID(ctx, mockDB.DB, id, schoolID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &id, &schoolID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.DeleteByExternalID(ctx, mockDB.DB, id, schoolID)
		assert.EqualError(t, err, fmt.Errorf("not found any quiz to delete: %w", pgx.ErrNoRows).Error())
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &id, &schoolID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.DeleteByExternalID(ctx, mockDB.DB, id, schoolID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "quizzes")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "status")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"external_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"school_id":   {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}

func TestQuizRepo_GetByQuestionGroupID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	pgQuestionGroupID := database.Text("group-id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgQuestionGroupID,
		)

		results, err := r.GetByQuestionGroupID(ctx, mockDB.DB, pgQuestionGroupID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgQuestionGroupID,
		)

		e := &entities.Quiz{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		results, err := r.GetByQuestionGroupID(ctx, mockDB.DB, pgQuestionGroupID)
		assert.Nil(t, err)
		assert.Equal(t, entities.Quizzes{e}, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestQuizRepo_DeleteByQuestionGroupID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizRepoWithSqlMock()

	id := database.Text("group-id")

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.DeleteByQuestionGroupID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.DeleteByQuestionGroupID(ctx, mockDB.DB, id)
		assert.EqualError(t, err, fmt.Errorf("quiz not found: %w", pgx.ErrNoRows).Error())
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.DeleteByQuestionGroupID(ctx, mockDB.DB, id)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "quizzes")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "status")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"question_group_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func QuizSetRepoWithSqlMock() (*QuizSetRepo, *testutil.MockDB) {
	r := &QuizSetRepo{}
	return r, testutil.NewMockDB()
}

func TestQuizset_Search(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()

	filter := QuizSetFilter{
		ObjectiveIDs: database.TextArray([]string{"objectID"}),
		Status:       database.Text("status"),
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&filter.ObjectiveIDs,
			&filter.Status,
		)

		results, err := r.Search(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&filter.ObjectiveIDs,
			&filter.Status,
		)

		e := &entities.QuizSet{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		results, err := r.Search(ctx, mockDB.DB, filter)
		assert.Nil(t, err)
		assert.Equal(t, entities.QuizSets{e}, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"lo_id":      {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"status":     {HasNullTest: true, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}

func TestQuizsetReppo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities.QuizSet{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities.QuizSet{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Equal(t, fmt.Errorf("can not create quizset"), err)
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.QuizSet{}
		fields, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestQuizsetReppo_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()

	id := database.Text("id")
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Delete(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Delete(ctx, mockDB.DB, id)
		assert.EqualError(t, err, fmt.Errorf("not found any quizset to delete: %w", pgx.ErrNoRows).Error())
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Delete(ctx, mockDB.DB, id)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "quiz_sets")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "status")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"quiz_set_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestQuizset_GetQuizSetByLoID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()
	loID := database.Text("loID")
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			loID,
		)

		results, err := r.GetQuizSetByLoID(ctx, mockDB.DB, loID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			loID,
		)

		e := &entities.QuizSet{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanFields(nil, fields, values)

		results, err := r.GetQuizSetByLoID(ctx, mockDB.DB, loID)
		assert.Nil(t, err)
		assert.Equal(t, e, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"lo_id":      {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestQuizsetReppo_GetQuizSetsContainQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()

	id := database.Text("quiz_id")

	t.Run("no rows affected", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), id)

		result, err := r.GetQuizSetsContainQuiz(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), id)
		quizSet := entities.QuizSet{}
		fields, values := quizSet.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		result, err := r.GetQuizSetsContainQuiz(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		assert.Equal(t, entities.QuizSets{&quizSet}, result)
	})
}

func TestQuizsetReppo_GetQuizSetsOfLOContainQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()

	loID := database.Text("lo_id")
	id := database.Text("quiz_id")

	t.Run("no rows affected", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), loID, id)

		result, err := r.GetQuizSetsOfLOContainQuiz(ctx, mockDB.DB, loID, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), loID, id)
		quizSet := entities.QuizSet{}
		fields, values := quizSet.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		result, err := r.GetQuizSetsOfLOContainQuiz(ctx, mockDB.DB, loID, id)
		assert.Nil(t, err)
		assert.Equal(t, entities.QuizSets{&quizSet}, result)
	})
}

func TestQuizsetRepo_GetTotalQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()
	ids := database.TextArray([]string{"id1", "id2", "id3"})
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), ids)

		_, err := r.GetTotalQuiz(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), ids)

		fields := []string{"total_quiz_los"}
		totals := make(map[string]int32)
		totals["id1"] = 1
		totals["id2"] = 2
		totals["id3"] = 3
		values := []interface{}{&totals}

		mockDB.MockScanFields(nil, fields, values)

		_, err := r.GetTotalQuiz(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}

func TestQuizset_GetQuizExternalIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()
	loID := database.Text("loID")
	limit := database.Int8(3)
	offset := database.Int8(0)
	t.Run("err select", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), loID, limit, offset).Once().Return(nil, puddle.ErrClosedPool)
		results, err := r.GetQuizExternalIDs(ctx, mockDB.DB, loID, limit, offset)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), loID, limit, offset).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Close").Once().Return(nil)
		mockDB.Rows.On("Next").Once().Return(true)
		for i := 0; i < int(limit.Int); i++ {
			mockDB.Rows.On("Scan", mock.Anything).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		_, err := r.GetQuizExternalIDs(ctx, mockDB.DB, loID, limit, offset)
		assert.Nil(t, err)
	})
}

func TestShuffledQuizset_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	t.Run("err insert", func(t *testing.T) {
		e := &entities.ShuffledQuizSet{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		result, err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result.Get())
	})

	t.Run("success with insert all fields", func(t *testing.T) {
		e := &entities.ShuffledQuizSet{}
		fields, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		result, err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)
		assert.Equal(t, e.ID, result)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestShuffledQuizset_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	id := database.Text("shuffled quiz set id")
	from := int64(1)
	to := int64(1)
	pgFrom := database.Int8(from)
	pgTo := database.Int8(to)
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			id,
			from,
			to,
		)

		results, err := r.Get(ctx, mockDB.DB, id, pgFrom, pgTo)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			id,
			from,
			to,
		)
		e := &entities.ShuffledQuizSet{}
		fields, values := e.FieldMap()
		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanFields(nil, fields, values)

		result, err := r.Get(ctx, mockDB.DB, id, pgFrom, pgTo)
		assert.Nil(t, err)
		assert.Equal(t, e, result)
	})
}

func TestShuffledQuizset_GetSeed(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	id := database.Text("shuffled quiz set id")
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), id)
		e := entities.ShuffledQuizSet{}

		fields := []string{"random_seed"}
		values := []interface{}{&e.RandomSeed}

		mockDB.MockRowScanFields(puddle.ErrClosedPool, fields, values)

		results, err := r.GetSeed(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results.Get())
	})
	// seed := database.Text("128757392934848")
	t.Run("success with select random seed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), id)

		e := entities.ShuffledQuizSet{}

		fields := []string{"random_seed"}
		values := []interface{}{&e.RandomSeed}

		seed := &e.RandomSeed
		mockDB.MockRowScanFields(nil, fields, values)

		result, err := r.GetSeed(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		assert.Equal(t, seed, &result)
	})
}

func TestShuffledQuizset_UpdateSubmissionHistory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	id := database.Text("shuffled quiz set id")
	ans := database.JSONB("answer log of student")
	t.Run("err update submission_history", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, ans, id)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)
		err := r.UpdateSubmissionHistory(ctx, mockDB.DB, id, ans)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with select random seed", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, ans, id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.UpdateSubmissionHistory(ctx, mockDB.DB, id, ans)
		assert.Nil(t, err)
	})
}

func TestShuffledQuizset_UpdateTotalCorrect(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	id := database.Text("shuffled quiz set id")
	t.Run("err update total correct", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, id)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)
		err := r.UpdateTotalCorrectness(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with update total correct", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.UpdateTotalCorrectness(ctx, mockDB.DB, id)
		assert.Nil(t, err)
	})
}

func TestShuffledQuizset_GetByStudyPlanItems(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	studyPlanItemID := database.TextArray([]string{"study_plan_item_id"})
	t.Run("err get shuffled quiz sets by study_plan_item_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			studyPlanItemID,
		)
		_, err := r.GetByStudyPlanItems(ctx, mockDB.DB, studyPlanItemID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with get shuffled quiz sets by studyPlanItemID", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			studyPlanItemID,
		)
		e := &entities.ShuffledQuizSet{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		results, err := r.GetByStudyPlanItems(ctx, mockDB.DB, studyPlanItemID)
		assert.Nil(t, err)
		assert.Equal(t, entities.ShuffledQuizSets{e}, results)
	})
}

func TestShuffledQuizset_GetBySessionID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	sessionID := database.Text("session_id")
	t.Run("err get shuffled quiz sets by sessionID", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			sessionID,
		)
		_, err := r.GetBySessionID(ctx, mockDB.DB, sessionID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with get shuffled quiz sets by sessionID", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			sessionID,
		)
		e := &entities.ShuffledQuizSet{}
		fields, values := e.FieldMap()

		mockDB.MockScanFields(nil, fields, values)
		results, err := r.GetBySessionID(ctx, mockDB.DB, sessionID)
		assert.Nil(t, err)
		assert.Equal(t, e, results)
	})
}

func TestShuffledQuizset_GetSubmissionHistory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")
	limit := database.Int4(3)
	offset := database.Int4(0)
	t.Run("err get shuffled quiz sets by study_plan_item_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
			limit.Get(),
			offset.Get(),
		)

		e := entities.ShuffledQuizSet{}
		e.QuizExternalIDs.Elements = append(e.QuizExternalIDs.Elements, database.Text("quiz_external_id_1"))
		fields := []string{"submission_history", "quiz_id"}
		values := []interface{}{&e.SubmissionHistory, &e.QuizExternalIDs.Elements[0]}

		mockDB.MockScanFields(nil, fields, values)
		_, _, err := r.GetSubmissionHistory(ctx, mockDB.DB, shuffledQuizSetID, limit, offset)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with get shuffled quiz sets by studyPlanItemID", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
			limit.Get(),
			offset.Get(),
		)

		e := entities.ShuffledQuizSet{}
		e.QuizExternalIDs.Elements = append(e.QuizExternalIDs.Elements, database.Text("quiz_external_id_1"))
		fields := []string{"submission_history", "quiz_id"}
		values := []interface{}{database.JSONB(nil), &e.QuizExternalIDs.Elements[0]}
		mockDB.MockScanFields(nil, fields, values)
		result, _, err := r.GetSubmissionHistory(ctx, mockDB.DB, shuffledQuizSetID, limit, offset)
		assert.Nil(t, err)
		expectRes := make(map[pgtype.Text]pgtype.JSONB)
		expectRes[e.QuizExternalIDs.Elements[0]] = e.SubmissionHistory
		assert.Equal(t, expectRes, result)
	})
}

func TestShuffledQuizset_GetStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")
	t.Run("err get shuffled quiz sets by study_plan_item_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
		)

		_, err := r.GetStudentID(ctx, mockDB.DB, shuffledQuizSetID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
		)

		expectStudentID := database.Text("stud011029")
		fields := []string{"student_id"}
		values := []interface{}{&expectStudentID}
		mockDB.MockScanFields(nil, fields, values)

		result, err := r.GetStudentID(ctx, mockDB.DB, shuffledQuizSetID)
		assert.True(t, errors.Is(err, nil))
		assert.Equal(t, expectStudentID, result)
	})
}

func TestShuffledQuizset_GetLoID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")
	t.Run("err get shuffled quiz sets by study_plan_item_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
		)

		_, err := r.GetLoID(ctx, mockDB.DB, shuffledQuizSetID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
		)

		expectLoID := database.Text("loID011029")
		fields := []string{"student_id"}
		values := []interface{}{&expectLoID}
		mockDB.MockScanFields(nil, fields, values)

		result, err := r.GetLoID(ctx, mockDB.DB, shuffledQuizSetID)
		assert.True(t, errors.Is(err, nil))
		assert.Equal(t, expectLoID, result)
	})
}

func TestShuffledQuizSetRepo_GetScore(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()

	ID := idutil.ULIDNow()
	shuffleQuizID := database.Text(ID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, shuffleQuizID)
		_, _, err := r.GetScore(ctx, mockDB.DB, shuffleQuizID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffleQuizID)
		var totalCorrectness, arrayLength pgtype.Int4
		mockDB.MockScanArray(nil, []string{"total_correctness", "array_length"}, [][]interface{}{
			{
				&totalCorrectness, &arrayLength,
			},
		})
		_, _, err := r.GetScore(ctx, mockDB.DB, shuffleQuizID)
		assert.Nil(t, err)
	})
}

func TestShuffledQuizset_IsFinishedQuizTest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")
	t.Run("err get shuffled quiz sets by study_plan_item_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
		)

		_, err := r.IsFinishedQuizTest(ctx, mockDB.DB, shuffledQuizSetID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			shuffledQuizSetID,
		)

		expectLoID := database.Bool(true)
		fields := []string{"is_finished"}
		values := []interface{}{&expectLoID}
		mockDB.MockScanFields(nil, fields, values)

		result, err := r.IsFinishedQuizTest(ctx, mockDB.DB, shuffledQuizSetID)
		assert.True(t, errors.Is(err, nil))
		assert.Equal(t, expectLoID, result)
	})
}

func TestShuffledQuizset_GetQuizIdx(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")
	quizID := database.Text("quiz_id")
	t.Run("err get shuffled quiz sets by study_plan_item_id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			quizID,
			shuffledQuizSetID,
		)

		_, err := r.GetQuizIdx(ctx, mockDB.DB, shuffledQuizSetID, quizID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			quizID,
			shuffledQuizSetID,
		)

		expectQuizIdx := database.Int4(1)
		fields := []string{"value"}
		values := []interface{}{&expectQuizIdx}
		mockDB.MockScanFields(nil, fields, values)

		result, err := r.GetQuizIdx(ctx, mockDB.DB, shuffledQuizSetID, quizID)
		assert.True(t, errors.Is(err, nil))
		assert.Equal(t, expectQuizIdx, result)
	})
}

func TestQuizRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	row := &mock_database.Row{}
	quizRepo := &QuizRepo{}
	e := &entities.Quiz{}
	_, values := e.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.Quiz{
				{
					ID:              database.Text("this is test"),
					ExternalID:      database.Text("this is test"),
					Country:         database.Text("this is test"),
					SchoolID:        database.Int4(10),
					LoIDs:           database.TextArray([]string{"this is test"}),
					Kind:            database.Text("this is test"),
					Question:        database.JSONB("{}"),
					Explanation:     database.JSONB("{}"),
					Options:         database.JSONB("{}"),
					TaggedLOs:       database.TextArray([]string{"this is test"}),
					DifficultLevel:  database.Int4(10),
					CreatedBy:       database.Text("this is test"),
					ApprovedBy:      database.Text("this is test"),
					Status:          database.Text("this is test"),
					UpdatedAt:       database.Timestamptz(time.Now()),
					CreatedAt:       database.Timestamptz(time.Now()),
					QuestionGroupID: database.Text("this is test"),
					LabelType:       database.Text("this is label type for test"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("QueryRow").Once().Return(row, nil)
				row.On("Scan", values...).Once().Return(nil)
				batchResults.On("Close").Once().Return(nil)
			},
		}, {
			name: "error send batch",
			req: []*entities.Quiz{
				{
					ID:              database.Text("this is test"),
					ExternalID:      database.Text("this is test"),
					Country:         database.Text("this is test"),
					SchoolID:        database.Int4(10),
					LoIDs:           database.TextArray([]string{"this is test"}),
					Kind:            database.Text("this is test"),
					Question:        database.JSONB("{}"),
					Explanation:     database.JSONB("{}"),
					Options:         database.JSONB("{}"),
					TaggedLOs:       database.TextArray([]string{"this is test"}),
					DifficultLevel:  database.Int4(10),
					CreatedBy:       database.Text("this is test"),
					ApprovedBy:      database.Text("this is test"),
					Status:          database.Text("this is test"),
					UpdatedAt:       database.Timestamptz(time.Now()),
					CreatedAt:       database.Timestamptz(time.Now()),
					QuestionGroupID: database.Text("this is test"),
				},
				{
					ID:              database.Text("this is test"),
					ExternalID:      database.Text("this is test"),
					Country:         database.Text("this is test"),
					SchoolID:        database.Int4(10),
					LoIDs:           database.TextArray([]string{"this is test"}),
					Kind:            database.Text("this is test"),
					Question:        database.JSONB("{}"),
					Explanation:     database.JSONB("{}"),
					Options:         database.JSONB("{}"),
					TaggedLOs:       database.TextArray([]string{"this is test"}),
					DifficultLevel:  database.Int4(10),
					CreatedBy:       database.Text("this is test"),
					ApprovedBy:      database.Text("this is test"),
					Status:          database.Text("this is test"),
					UpdatedAt:       database.Timestamptz(time.Now()),
					CreatedAt:       database.Timestamptz(time.Now()),
					QuestionGroupID: database.Text("this is test"),
				},
				{
					ID:              database.Text("this is test"),
					ExternalID:      database.Text("this is test"),
					Country:         database.Text("this is test"),
					SchoolID:        database.Int4(10),
					LoIDs:           database.TextArray([]string{"this is test"}),
					Kind:            database.Text("this is test"),
					Question:        database.JSONB("{}"),
					Explanation:     database.JSONB("{}"),
					Options:         database.JSONB("{}"),
					TaggedLOs:       database.TextArray([]string{"this is test"}),
					DifficultLevel:  database.Int4(10),
					CreatedBy:       database.Text("this is test"),
					ApprovedBy:      database.Text("this is test"),
					Status:          database.Text("this is test"),
					UpdatedAt:       database.Timestamptz(time.Now()),
					CreatedAt:       database.Timestamptz(time.Now()),
					QuestionGroupID: database.Text("this is test"),
				},
				{
					ID:              database.Text("this is test"),
					ExternalID:      database.Text("this is test"),
					Country:         database.Text("this is test"),
					SchoolID:        database.Int4(10),
					LoIDs:           database.TextArray([]string{"this is test"}),
					Kind:            database.Text("this is test"),
					Question:        database.JSONB("{}"),
					Explanation:     database.JSONB("{}"),
					Options:         database.JSONB("{}"),
					TaggedLOs:       database.TextArray([]string{"this is test"}),
					DifficultLevel:  database.Int4(10),
					CreatedBy:       database.Text("this is test"),
					ApprovedBy:      database.Text("this is test"),
					Status:          database.Text("this is test"),
					UpdatedAt:       database.Timestamptz(time.Now()),
					CreatedAt:       database.Timestamptz(time.Now()),
					QuestionGroupID: database.Text("this is test"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.QueryRow: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				row := &mock_database.Row{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(pgx.ErrTxClosed)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		_, err := quizRepo.Upsert(
			ctx,
			db,
			testCase.req.([]*entities.Quiz),
		)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestShuffledQuizSetRepo_Retrieve(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()
	ids := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, ids)

		shuffleQuizzes, err := r.Retrieve(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, shuffleQuizzes)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, ids)

		e := &entities.ShuffledQuizSet{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(idutil.ULIDNow())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.Retrieve(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestShuffledQuizSetRepo_GetExternalIDsFromSubmissionHistory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ShuffledQuizSetRepoWithSQLMock()

	ID := idutil.ULIDNow()
	shuffleQuizID := database.Text(ID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, shuffleQuizID)
		shuffleQuizzes, err := r.GetExternalIDsFromSubmissionHistory(ctx, mockDB.DB, shuffleQuizID, true)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Empty(t, shuffleQuizzes)
	})
	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffleQuizID)

		e := &entities.ShuffledQuizSet{}
		// fields, values := e.FieldMap()
		val := pgtype.TextArray{}
		_ = e.ID.Set(idutil.ULIDNow())
		mockDB.MockScanArray(nil, []string{"coalese"}, [][]interface{}{
			{
				&val,
			},
		})
		_, err := r.GetExternalIDsFromSubmissionHistory(ctx, mockDB.DB, shuffleQuizID, true)
		assert.Nil(t, err)
	})
}

func TestShuffledQuizSetRepo_ListExternalIDsFromSubmissionHistory(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &ShuffledQuizSetRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"study_plan_item_id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			req:         database.TextArray([]string{"study_plan_item_id"}),
			expectedErr: fmt.Errorf("ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Err").Once().Return(nil)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			req:         database.TextArray([]string{"study_plan_item_id"}),
			expectedErr: fmt.Errorf("ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				rows.On("Err").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.ListExternalIDsFromSubmissionHistory(ctx, db, testCase.req.(pgtype.TextArray), false)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestQuizRepo_CountQuizOnLO(t *testing.T) {
	t.Parallel()
	r, mockDB := QuizSetRepoWithSqlMock()
	pgLOIDs := database.TextArray([]string{"loIDS"})

	testCases := []TestCase{
		{
			name: "err select",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
					mock.AnythingOfType("string"),
					&pgLOIDs,
					mock.AnythingOfType("string"),
				)
			},
			expectedErr: fmt.Errorf("repo.DB.QueryEx: %w", puddle.ErrClosedPool),
		},
		{
			name: "success with count quiz on lo",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.AnythingOfType("string"),
					&pgLOIDs,
					mock.AnythingOfType("string"),
				)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.AnythingOfType("*pgtype.Text"), mock.AnythingOfType("*pgtype.Int4")).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.CountQuizOnLO(ctx, mockDB.DB, pgLOIDs)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestCalculateHighestSubmissionScore(t *testing.T) {
	t.Parallel()

	repo, mockDB := ShuffledQuizSetRepoWithSQLMock()

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"study_plan_item_id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
			},
		},
		{
			name:        "query error",
			req:         database.TextArray([]string{"study_plan_item_id"}),
			expectedErr: status.Error(codes.Internal, fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Query: %w", fmt.Errorf("error")).Error()),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, fmt.Errorf("error"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			req:         database.TextArray([]string{"study_plan_item_id"}),
			expectedErr: fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.CalculateHighestSubmissionScore(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestShuffledQuizSetRepo_GenerateExamLOSubmission(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}
	shuffledQuizSetID := database.Text("shuffledQuizSetID-1")

	e := &entities.ExamLOSubmission{}
	fields, values := e.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffledQuizSetID)
				mockDB.MockScanFields(nil, fields, values)
			},
			req:          shuffledQuizSetID,
			expectedResp: &entities.ExamLOSubmission{},
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffledQuizSetID)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          shuffledQuizSetID,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GenerateExamLOSubmission(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_GetExternalIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}
	shuffledQuizSetID := database.Text("shuffledQuizSetID-1")
	var result pgtype.TextArray

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffledQuizSetID)
				mockDB.MockScanFields(nil, []string{"quiz_external_ids"}, []interface{}{&result})
			},
			req:          shuffledQuizSetID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffledQuizSetID)
				mockDB.MockScanFields(pgx.ErrNoRows, []string{"quiz_external_ids"}, []interface{}{&result})
			},
			req:          shuffledQuizSetID,
			expectedResp: pgtype.TextArray{},
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetExternalIDs(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_GetByStudyPlanItemIdentities(t *testing.T) {
	t.Parallel()

	repo, mockDB := ShuffledQuizSetRepoWithSQLMock()
	e := &entities.ShuffledQuizSet{}
	fields, values := e.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*StudyPlanItemIdentity{
				{StudentID: database.Text("11"), StudyPlanID: database.Text("12"), LearningMaterialID: database.Text("13")},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name: "query error",
			req: []*StudyPlanItemIdentity{
				{StudentID: database.Text("11"), StudyPlanID: database.Text("12"), LearningMaterialID: database.Text("13")},
			},
			expectedErr: fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.setup(ctx)

		_, err := repo.GetByStudyPlanItemIdentities(ctx, mockDB.DB, tc.req.([]*StudyPlanItemIdentity))
		assert.Equal(t, tc.expectedErr, err)
	}
}

func TestQuizRepo_GetQuizByExternalID(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &QuizRepo{}
	externalID := database.Text("external_id")

	result := &entities.Quiz{}
	fields, values := result.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, externalID)
				mockDB.MockScanFields(nil, fields, values)
			},
			req:          externalID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, externalID)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          externalID,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetQuizByExternalID(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_GetCorrectnessInfo(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")
	externalID := database.Text("external_id")

	result := &entities.CorrectnessInfo{}
	fields, values := result.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffledQuizSetID, externalID)
				mockDB.MockScanFields(nil, fields, values)
			},
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffledQuizSetID, externalID)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetCorrectnessInfo(ctx, mockDB.DB, shuffledQuizSetID, externalID)
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_UpdateTotalCorrectnessAndSubmissionHistory(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}
	shuffledQuiSet := &entities.ShuffledQuizSet{
		ID:                database.Text("shuffled_quiz_set_id"),
		TotalCorrectness:  database.Int4(1),
		SubmissionHistory: database.JSONB("answer log of student"),
		UpdatedAt:         database.Timestamptz(time.Now()),
	}

	stmt := `
    UPDATE shuffled_quiz_sets
       SET total_correctness = $1,
           submission_history = submission_history || $2,
           updated_at = $3
     WHERE shuffled_quiz_set_id = $4;
	`

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, stmt, shuffledQuiSet.TotalCorrectness, shuffledQuiSet.SubmissionHistory, shuffledQuiSet.UpdatedAt, shuffledQuiSet.ID)
			},
			req:         shuffledQuiSet,
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrNoRows, mock.Anything, stmt, shuffledQuiSet.TotalCorrectness, shuffledQuiSet.SubmissionHistory, shuffledQuiSet.UpdatedAt, shuffledQuiSet.ID)
			},
			req:         shuffledQuiSet,
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.UpdateTotalCorrectnessAndSubmissionHistory(ctx, mockDB.DB, testCase.req.(*entities.ShuffledQuizSet))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestShuffledQuizSetRepo_InsertLOSubmission(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")

	result := &entities.LOSubmissionAnswerKey{}
	fields, values := result.FieldMap()

	stmt := `
    INSERT INTO lo_submission (
        submission_id,
        student_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        created_at,
        updated_at,
        deleted_at,
        total_point
    )
    SELECT generate_ulid() AS submission_id,
           SQ.student_id,
           SQ.study_plan_id,
           SQ.learning_material_id,
           SQ.shuffled_quiz_set_id,
           SQ.created_at,
           SQ.updated_at,
           SQ.deleted_at,
           COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(SQ.quiz_external_ids)), 0)::INT AS total_point
    FROM shuffled_quiz_sets SQ
    WHERE shuffled_quiz_set_id = $1::TEXT
    ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
        updated_at = EXCLUDED.updated_at,
        total_point = EXCLUDED.total_point
    RETURNING student_id,submission_id,study_plan_id,learning_material_id,shuffled_quiz_set_id;
	`

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, stmt, shuffledQuizSetID)
				mockDB.MockRowScanFields(nil, fields, values)
			},
			req:          shuffledQuizSetID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, stmt, shuffledQuizSetID)
				mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          shuffledQuizSetID,
			expectedResp: result,
			expectedErr:  fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.UpsertLOSubmission(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_InsertFlashCardSubmission(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}
	shuffledQuizSetID := database.Text("shuffled_quiz_set_id")

	result := &entities.FlashCardSubmissionAnswerKey{}
	fields, values := result.FieldMap()

	stmt := `
    INSERT INTO flash_card_submission (
        submission_id,
        student_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        created_at,
        updated_at,
        deleted_at,
        total_point
    )
    SELECT generate_ulid() AS submission_id,
           SQ.student_id,
           SQ.study_plan_id,
           SQ.learning_material_id,
           SQ.shuffled_quiz_set_id,
           SQ.created_at,
           SQ.updated_at,
           SQ.deleted_at,
           COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(SQ.quiz_external_ids)), 0)::INT AS total_point
    FROM shuffled_quiz_sets SQ
    WHERE shuffled_quiz_set_id = $1::TEXT
    ON CONFLICT ON CONSTRAINT flash_card_submission_shuffled_quiz_set_id_un DO UPDATE SET
        updated_at = EXCLUDED.updated_at,
        total_point = EXCLUDED.total_point
    RETURNING student_id,submission_id,study_plan_id,learning_material_id,shuffled_quiz_set_id;
	`

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, stmt, shuffledQuizSetID)
				mockDB.MockRowScanFields(nil, fields, values)
			},
			req:          shuffledQuizSetID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, stmt, shuffledQuizSetID)
				mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          shuffledQuizSetID,
			expectedResp: result,
			expectedErr:  fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.UpsertFlashCardSubmission(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_GetRelatedLearningMaterial(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ShuffledQuizSetRepo{}

	e := &entities.LearningMaterial{}
	testCases := []TestCase{
		{
			name: "happy case",
			req:  database.Text("set-1"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text("set-1"))

				fields, values := e.FieldMap()
				mockDB.MockRowScanFields(nil, fields, values)
			},
			expectedResp: e,
		},
		{
			name: "missing learning material",
			req:  database.Text("set-1"),
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text("set-1"))
				fields, values := e.FieldMap()
				mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
			},
			expectedErr: fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetRelatedLearningMaterial(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestShuffledQuizSetRepo_GetTotalPointsByQuizSetID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuizSetRepoWithSqlMock()

	ID := idutil.ULIDNow()
	shuffleQuizID := database.Text(ID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, shuffleQuizID)
		_, err := r.GetTotalPointsByQuizSetID(ctx, mockDB.DB, shuffleQuizID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, shuffleQuizID)
		var totalPoints pgtype.Int4
		mockDB.MockScanArray(nil, []string{"total_points"}, [][]interface{}{
			{
				&totalPoints,
			},
		})
		_, err := r.GetTotalPointsByQuizSetID(ctx, mockDB.DB, shuffleQuizID)
		assert.Nil(t, err)
	})
}
