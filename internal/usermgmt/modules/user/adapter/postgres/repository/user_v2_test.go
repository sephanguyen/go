package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserV2WithSqlMock() (*UserRepoV2, *testutil.MockDB) {
	r := &UserRepoV2{}
	return r, testutil.NewMockDB()
}

func TestUserRepoV2_GetByAuthInfo(t *testing.T) {
	t.Parallel()

	userID := "userID"
	projectID := "projectID"
	tenantID := "tenantID"

	t.Run("failed to get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UserV2WithSqlMock()
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&userID,
			&projectID,
			&tenantID,
		)

		groups, err := r.GetByAuthInfo(ctx, mockDB.DB, "", userID, projectID, tenantID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, groups)
	})

	t.Run("happy case", func(tt *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UserV2WithSqlMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&userID,
			&projectID,
			&tenantID,
		)

		e := &entity.LegacyUser{}

		fields, values := e.FieldMap()

		e.ID.Set(idutil.ULIDNow())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		user, err := r.GetByAuthInfo(ctx, mockDB.DB, "", userID, projectID, tenantID)
		assert.NoError(tt, err)
		assert.Equal(tt, e, user)
	})
}

func TestUserRepoV2_GetByAuthInfoV2(t *testing.T) {
	t.Parallel()

	userID := "userID"
	projectID := "projectID"
	tenantID := "tenantID"

	t.Run("failed to get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UserV2WithSqlMock()
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&userID,
			&projectID,
			&tenantID,
		)

		groups, err := r.GetByAuthInfoV2(ctx, mockDB.DB, "", userID, projectID, tenantID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, groups)
	})

	t.Run("happy case", func(tt *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UserV2WithSqlMock()
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&userID,
			&projectID,
			&tenantID,
		)

		e := &entity.AuthUser{}

		fields, values := e.FieldMap()

		e.UserID.Set(idutil.ULIDNow())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		user, err := r.GetByAuthInfoV2(ctx, mockDB.DB, "", userID, projectID, tenantID)
		assert.NoError(tt, err)
		assert.Equal(tt, e, user)
	})
}

func TestUserRepoV2_GetByUsername(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userRepo := &UserRepoV2{}
	mockDB := testutil.NewMockDB()
	rows := &mock_database.Rows{}
	authUser := &entity.AuthUser{}
	_, scanFields := authUser.FieldMap()

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:      "query return no rows err",
			expectErr: pgx.ErrNoRows,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	for index, testcase := range testCases {
		t.Run(fmt.Sprintf("%s-%d", testcase.name, index), func(t *testing.T) {
			testcase.setup()
			role, err := userRepo.GetByUsername(ctx, mockDB.DB, "username", "manabie")
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			} else {
				assert.NotNil(t, role)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, rows)
		})
	}
}

func TestUserRepoV2_GetByEmail(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userRepo := &UserRepoV2{}
	mockDB := testutil.NewMockDB()
	rows := &mock_database.Rows{}
	authUser := &entity.AuthUser{}
	_, scanFields := authUser.FieldMap()

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			name:      "query return no rows err",
			expectErr: pgx.ErrNoRows,
			setup: func() {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", scanFields...).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	for index, testcase := range testCases {
		t.Run(fmt.Sprintf("%s-%d", testcase.name, index), func(t *testing.T) {
			testcase.setup()
			role, err := userRepo.GetByEmail(ctx, mockDB.DB, "email", "manabie")
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			} else {
				assert.NotNil(t, role)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, rows)
		})
	}
}
