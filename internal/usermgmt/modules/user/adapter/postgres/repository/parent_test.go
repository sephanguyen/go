package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ParentRepoWithSqlMock() (*ParentRepo, *testutil.MockDB) {
	repo := &ParentRepo{}
	return repo, testutil.NewMockDB()
}

func TestParentRepo_Create(t *testing.T) {
	t.Parallel()
	now := time.Now()
	userGroup := entity.UserGroup{}
	_, userGroupValues := userGroup.FieldMap()
	e := &entity.Parent{}
	parentIDStr := uuid.NewString()
	_ = e.ID.Set(parentIDStr)
	_ = e.UpdatedAt.Set(now)
	_ = e.CreatedAt.Set(now)
	_ = e.LegacyUser.ID.Set(parentIDStr)
	_ = e.LegacyUser.Group.Set(entity.UserGroupParent)
	_ = e.LegacyUser.UpdatedAt.Set(now)
	_ = e.LegacyUser.CreatedAt.Set(now)
	_ = e.LegacyUser.DeviceToken.Set(nil)
	_ = e.LegacyUser.AllowNotification.Set(true)
	_ = e.LegacyUser.ResourcePath.Set(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		user := &e.LegacyUser
		cmdTag := pgconn.CommandTag([]byte(`1`))

		_, userValues := user.FieldMap()
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)

		_, parentValues := e.FieldMap()
		argsParent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(parentValues))...)
		mockDB.DB.On("Exec", argsParent...).Once().Return(cmdTag, nil)

		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		mockDB.DB.On("Exec", argsUserGroup...).Return(cmdTag, nil).Once()

		err := parentRepo.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("Insert user failed", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		user := &e.LegacyUser
		_, userValues := user.FieldMap()
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(nil, pgx.ErrTxClosed)

		err := parentRepo.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, errors.Wrap(pgx.ErrTxClosed, "Insert() user_id"), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("Insert parent failed", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		user := &e.LegacyUser
		cmdTag := pgconn.CommandTag([]byte(`1`))

		_, userValues := user.FieldMap()
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)

		_, parentValues := e.FieldMap()
		argsParent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(parentValues))...)
		mockDB.DB.On("Exec", argsParent...).Once().Return(nil, pgx.ErrTxClosed)

		err := parentRepo.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, errors.Wrap(pgx.ErrTxClosed, "Insert() parent_id"), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("Insert user group failed", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		user := &e.LegacyUser
		cmdTag := pgconn.CommandTag([]byte(`1`))

		_, userValues := user.FieldMap()
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)

		_, parentValues := e.FieldMap()
		argsParent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(parentValues))...)
		mockDB.DB.On("Exec", argsParent...).Once().Return(cmdTag, nil)

		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		mockDB.DB.On("Exec", argsUserGroup...).Return(nil, pgx.ErrTxClosed)

		err := parentRepo.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, fmt.Errorf("err insert UserGroup: %w", pgx.ErrTxClosed), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("Insert user group with no row affected", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		user := &e.LegacyUser
		cmdTag1 := pgconn.CommandTag([]byte(`1`))

		_, userValues := user.FieldMap()
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag1, nil)

		_, parentValues := e.FieldMap()
		argsParent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(parentValues))...)
		mockDB.DB.On("Exec", argsParent...).Once().Return(cmdTag1, nil)

		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		cmdTag0 := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsUserGroup...).Return(cmdTag0, nil)

		err := parentRepo.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, fmt.Errorf("%d RowsAffected: %w", cmdTag0.RowsAffected(), ErrUnAffected), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}

func TestParentRepo_GetByIds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ids := pgtype.TextArray{}
	_ = ids.Set([]string{uuid.NewString()})

	_, userValues := (&entity.LegacyUser{}).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)
	_, parentValues := (&entity.Parent{}).FieldMap()
	argsParent := append([]interface{}{}, genSliceMock(len(parentValues))...)

	t.Run("success with select all fields", func(t *testing.T) {
		r, mockDB := ParentRepoWithSqlMock()

		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &ids).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(ids.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsUser, argsParent...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)

		parents, err := r.GetByIds(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.NotNil(t, parents)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("Query return error (except no rows error)", func(t *testing.T) {
		r, mockDB := ParentRepoWithSqlMock()

		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &ids).Once().Return(nil, pgx.ErrTxClosed)

		parents, err := r.GetByIds(ctx, mockDB.DB, ids)
		assert.EqualError(t, fmt.Errorf("db.Query: %w", pgx.ErrTxClosed), err.Error())
		assert.Nil(t, parents)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("Query return no rows error", func(t *testing.T) {
		r, mockDB := ParentRepoWithSqlMock()

		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &ids).Once().Return(nil, pgx.ErrNoRows)

		parents, err := r.GetByIds(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, entity.Parents{}, parents)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("Rows Scan failed", func(t *testing.T) {
		r, mockDB := ParentRepoWithSqlMock()

		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &ids).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", append(argsUser, argsParent...)...).Once().Return(pgx.ErrNoRows)
		mockDB.Rows.On("Close").Once().Return(nil)

		parents, err := r.GetByIds(ctx, mockDB.DB, ids)
		assert.EqualError(t, fmt.Errorf("row.Scan: %w", pgx.ErrNoRows), err.Error())
		assert.Nil(t, parents)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestParentRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	now := time.Now()

	e1 := &entity.Parent{}
	parent1IDStr := uuid.NewString()
	_ = e1.ID.Set(parent1IDStr)
	_ = e1.UpdatedAt.Set(now)
	_ = e1.CreatedAt.Set(now)
	_ = e1.LegacyUser.ID.Set(parent1IDStr)
	_ = e1.LegacyUser.Group.Set(entity.UserGroupParent)
	_ = e1.LegacyUser.UpdatedAt.Set(now)
	_ = e1.LegacyUser.CreatedAt.Set(now)
	_ = e1.LegacyUser.DeviceToken.Set(nil)
	_ = e1.LegacyUser.AllowNotification.Set(true)
	e2 := &entity.Parent{}
	parent2IDStr := uuid.NewString()
	_ = e2.ID.Set(parent2IDStr)
	_ = e2.UpdatedAt.Set(now)
	_ = e2.CreatedAt.Set(now)
	_ = e2.LegacyUser.ID.Set(parent2IDStr)
	_ = e2.LegacyUser.Group.Set(entity.UserGroupParent)
	_ = e2.LegacyUser.UpdatedAt.Set(now)
	_ = e2.LegacyUser.CreatedAt.Set(now)
	_ = e2.LegacyUser.DeviceToken.Set(nil)
	_ = e2.LegacyUser.AllowNotification.Set(true)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case (1 parent)", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entity.Parent{e1})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("happy case (2 parents)", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entity.Parent{e1, e2})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("batch Exec failed", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults.On("Exec").Return(cmdTag, pgx.ErrTxClosed)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entity.Parent{e1})
		assert.EqualError(t, errors.Wrap(pgx.ErrTxClosed, "batchResults.Exec"), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("batch Exec no rows affected", func(t *testing.T) {
		parentRepo, mockDB := ParentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`0`))
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entity.Parent{e1})
		assert.EqualError(t, fmt.Errorf("parent not inserted"), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}

func TestParentRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	parentID := database.Text("id")
	_, parentValues := (&entity.Parent{}).FieldMap()
	argsParent := append([]interface{}{}, genSliceMock(len(parentValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := ParentRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &parentID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsParent...).Once().Return(nil)
		parents, err := repo.GetByID(ctx, mockDB.DB, parentID)
		assert.Nil(t, err)
		assert.NotNil(t, parents)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := ParentRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &parentID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsParent...).Once().Return(puddle.ErrClosedPool)
		parent, err := repo.GetByID(ctx, mockDB.DB, parentID)
		assert.Nil(t, parent)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
