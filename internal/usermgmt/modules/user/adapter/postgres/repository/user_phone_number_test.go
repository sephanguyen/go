package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserPhoneNumberRepoWithSqlMock() (*UserPhoneNumberRepo, *testutil.MockDB) {
	r := &UserPhoneNumberRepo{}
	return r, testutil.NewMockDB()
}

func userPhoneNumberRepoWithMock() (*UserPhoneNumberRepo, *testutil.MockDB) {
	repo := &UserPhoneNumberRepo{}
	return repo, testutil.NewMockDB()
}

func TestUserPhoneNumberRepo_Upsert(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	userPhoneNumber := &UserPhoneNumberRepo{}

	stubUserPhoneNumber := &entity.UserPhoneNumber{
		ID:              pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
		UserID:          pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
		PhoneNumber:     pgtype.Text{String: "", Status: pgtype.Present},
		PhoneNumberType: pgtype.Text{String: entity.StudentPhoneNumber, Status: pgtype.Present},
		ResourcePath:    pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         []*entity.UserPhoneNumber{stubUserPhoneNumber},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "normal case with empty userPhoneNumber",
			req:         []*entity.UserPhoneNumber{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				// this case no need to mock any thing because it will return immediate nil err
			},
		},
		{
			name:        "error send batch",
			req:         []*entity.UserPhoneNumber{stubUserPhoneNumber},
			expectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := userPhoneNumber.Upsert(ctx, db, testCase.req.([]*entity.UserPhoneNumber))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserPhoneNumberRepo_FindByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := UserPhoneNumberRepoWithSqlMock()
	id := database.Text(idutil.ULIDNow())

	userPhoneNumber := &entity.UserPhoneNumber{}
	fields, values := userPhoneNumber.FieldMap()

	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr error
		setup       func(context.Context) *mock_database.Ext
	}{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)
			_, err := repo.FindByUserID(testCase.ctx, db, id)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_SoftDeleteByUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := userPhoneNumberRepoWithMock()

	userIDs := database.TextArray([]string{"userID-1", "userID-2"})

	t.Run("err exec", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &userIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.SoftDeleteByUserIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &userIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.SoftDeleteByUserIDs(ctx, mockDB.DB, userIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "user_phone_number")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}
