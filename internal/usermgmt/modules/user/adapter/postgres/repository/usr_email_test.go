package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UsrEmailWithSqlMock() (*UsrEmailRepo, *testutil.MockDB) {
	r := &UsrEmailRepo{}
	return r, testutil.NewMockDB()
}

func TestUsrEmailRepo_Create(t *testing.T) {
	t.Parallel()

	t.Run("failed to create usr_email", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UsrEmailWithSqlMock()

		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		usrEmail := &entity.UsrEmail{}
		database.AllNullEntity(usrEmail)
		fields, values := usrEmail.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		createdUsrEmail, err := r.Create(ctx, mockDB.DB, database.Text("example-usr-id"), database.Text("example-email"))

		assert.Nil(t, createdUsrEmail)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("happy case, create usr_email", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UsrEmailWithSqlMock()

		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		usrEmail := &entity.UsrEmail{}
		database.AllNullEntity(usrEmail)
		fields, values := usrEmail.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)

		createdUsrEmail, err := r.Create(ctx, mockDB.DB, database.Text("example-usr-id"), database.Text("example-email"))

		assert.Nil(t, err)
		assert.NotNil(t, createdUsrEmail)
	})
}

func TestUsrEmailRepo_CreateMultiple(t *testing.T) {
	t.Parallel()

	t.Run("failed to CreateMultiple usr_email", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UsrEmailWithSqlMock()

		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("QueryRow").Once().Return(mockDB.Row)
		batchResults.On("Close").Once().Return(nil)

		usrEmail := &entity.UsrEmail{}
		database.AllNullEntity(usrEmail)
		fields, values := usrEmail.FieldMap()
		users := []*entity.LegacyUser{
			{
				ID:    database.Text(idutil.ULIDNow()),
				Email: database.Text("student-01@example.com"),
			},
			{
				ID:    database.Text(idutil.ULIDNow()),
				Email: database.Text("student-02@example.com"),
			},
		}

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		createdUsrEmail, err := r.CreateMultiple(ctx, mockDB.DB, users)
		assert.Nil(t, createdUsrEmail)
		expectedErr := InternalError{
			RawError: errors.New("database.InsertReturning returns no row: no rows in result set"),
		}

		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("CreateMultiple usr_email successfully", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := UsrEmailWithSqlMock()

		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("QueryRow").Twice().Return(mockDB.Row)
		batchResults.On("Close").Once().Return(nil)

		usrEmail := &entity.UsrEmail{}
		database.AllNullEntity(usrEmail)
		fields, values := usrEmail.FieldMap()
		users := []*entity.LegacyUser{
			{
				ID:    database.Text(idutil.ULIDNow()),
				Email: database.Text("student-01@example.com"),
			},
			{
				ID:    database.Text(idutil.ULIDNow()),
				Email: database.Text("student-02@example.com"),
			},
		}

		mockDB.MockRowScanFields(nil, fields, values)
		mockDB.MockRowScanFields(nil, fields, values)

		createdUsrEmail, err := r.CreateMultiple(ctx, mockDB.DB, users)

		assert.Nil(t, err)
		assert.NotNil(t, createdUsrEmail)
	})
}

func TestUsrEmailRepo_UpdateEmail(t *testing.T) {
	t.Parallel()
	usrEmailRepo, db := UsrEmailWithSqlMock()
	type updateEmail struct {
		usrID    pgtype.Text
		oldEmail pgtype.Text
		newEmail pgtype.Text
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: updateEmail{
				usrID:    database.Text(idutil.ULIDNow()),
				oldEmail: database.Text("old-email.staff@example.com"),
				newEmail: database.Text("new-email.staff@example.com"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "not row affected",
			req: updateEmail{
				usrID:    database.Text(idutil.ULIDNow()),
				oldEmail: database.Text("old-email.staff@example.com"),
				newEmail: database.Text("new-email.staff@example.com"),
			},
			expectedErr: ErrNoRowAffected,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(failedTag))
				db.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: updateEmail{
				usrID:    database.Text(idutil.ULIDNow()),
				oldEmail: database.Text("old-email.staff@example.com"),
				newEmail: database.Text("new-email.staff@example.com"),
			},
			expectedErr: fmt.Errorf("db.Exec: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				db.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		userGroup := testCase.req.(updateEmail)
		err := usrEmailRepo.UpdateEmail(ctx, db.DB, userGroup.usrID, userGroup.oldEmail, userGroup.newEmail)
		assert.Equal(t, testCase.expectedErr, err)
	}
}
