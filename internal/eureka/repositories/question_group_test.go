package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func QuestionGroupRepoWithSqlMock() (*QuestionGroupRepo, *testutil.MockDB) {
	r := &QuestionGroupRepo{}
	return r, testutil.NewMockDB()
}

func TestQuestionGroupRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuestionGroupRepoWithSqlMock()

	t.Run("upsert successfully", func(t *testing.T) {
		e := entities.QuestionGroup{}
		fields, values := e.FieldMapUpsert()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		n, err := r.Upsert(ctx, mockDB.DB, &e)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, n)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("upsert failed", func(t *testing.T) {
		e := entities.QuestionGroup{}
		fields, values := e.FieldMapUpsert()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		n, err := r.Upsert(ctx, mockDB.DB, &e)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.EqualValues(t, 0, n)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := entities.QuestionGroup{
			QuestionGroupID: database.Text("id"),
		}
		fields, values := e.FieldMapUpsert()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		n, err := r.Upsert(ctx, mockDB.DB, &e)
		assert.NoError(t, err)
		assert.EqualValues(t, 0, n)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestQuestionGroupRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuestionGroupRepoWithSqlMock()

	groupID := "group-id"
	e := &entities.QuestionGroup{}
	selectFields, value := e.FieldMap()
	_ = e.QuestionGroupID.Set(groupID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &groupID)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		res, err := r.FindByID(ctx, mockDB.DB, groupID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, res)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &groupID)
		mockDB.MockRowScanFields(nil, selectFields, value)

		res, err := r.FindByID(ctx, mockDB.DB, groupID)

		assert.Nil(t, err)
		assert.Equal(t, &entities.QuestionGroup{QuestionGroupID: database.Text(groupID)}, res)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestQuestionGroupRepo_GetByQuestionGroupIDAndLoID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuestionGroupRepoWithSqlMock()

	groupID := "group-id"
	loID := "lo-id"

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything,
			mock.Anything,
		)

		res, err := r.GetByQuestionGroupIDAndLoID(ctx, mockDB.DB, database.Text(groupID), database.Text(loID))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, res)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.QuestionGroup{}
		selectFields, values := e.FieldMap()
		_ = e.LearningMaterialID.Set(loID)
		_ = e.QuestionGroupID.Set(groupID)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{values})

		res, err := r.GetByQuestionGroupIDAndLoID(ctx, mockDB.DB, database.Text(groupID), database.Text(loID))

		assert.Nil(t, err)
		assert.Equal(t, &entities.QuestionGroup{LearningMaterialID: database.Text(loID), QuestionGroupID: database.Text(groupID)}, res)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestQuestionGroupRepo_GetQuestionGroupsByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuestionGroupRepoWithSqlMock()
	questionGroupIDs := []string{"id-1", "id-2", "id-3"}
	e := &entities.QuestionGroup{}
	selectFields, _ := e.FieldMap()
	selectFields = append(selectFields, "total_children")
	selectFields = append(selectFields, "total_points")
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

	resMock := entities.QuestionGroups{
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("id-1"),
			LearningMaterialID: database.Text("lm-1"),
			Name:               database.Text("name 1"),
			Description:        database.Text("description 1"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("id-1"),
			LearningMaterialID: database.Text("lm-1"),
			Name:               database.Text("name 2"),
			Description:        database.Text("description 2"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("id-3"),
			LearningMaterialID: database.Text("lm-1"),
			Name:               database.Text("name 3"),
			Description:        database.Text("description 3"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
	}

	values := make([][]interface{}, 0, len(resMock))
	for i := range resMock {
		resMock[i].SetTotalChildrenAndPoints(3, 5)
		_, v := resMock[i].FieldMap()
		v = append(v, resMock[i].TotalChildren())
		v = append(v, resMock[i].TotalPoints())
		values = append(values, v)
	}

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.TextArray(questionGroupIDs))
		mockDB.MockScanArray(nil, selectFields, values)

		res, err := r.GetQuestionGroupsByIDs(ctx, mockDB.DB, questionGroupIDs...)

		assert.Nil(t, err)
		assert.EqualValues(t, resMock, res)
	})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.TextArray(questionGroupIDs))
		mockDB.MockScanArray(puddle.ErrClosedPool, selectFields, values)

		res, err := r.GetQuestionGroupsByIDs(ctx, mockDB.DB, questionGroupIDs...)

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, res)
	})
}

func TestQuestionGroupRepo_DeleteByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := QuestionGroupRepoWithSqlMock()

	id := database.Text("id")
	t.Run("err delete", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.DeleteByID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.DeleteByID(ctx, mockDB.DB, id)
		assert.EqualError(t, err, fmt.Errorf("question group not found: %w", pgx.ErrNoRows).Error())
	})

	t.Run("delete successfully", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &id)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.DeleteByID(ctx, mockDB.DB, id)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "question_group")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
	})
}
