package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainUserRepoWithSqlMock() (*DomainUserRepo, *testutil.MockDB) {
	r := &DomainUserRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainUserRepo_create(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainUserRepoWithSqlMock()
		user, err := NewUser(entity.EmptyUser{})
		if err != nil {
			t.Fatal(err)
		}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		cmdTag := pgconn.CommandTag(`1`)
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err = repo.create(ctx, mockDB.DB, user)
		assert.Nil(t, err)
	})
	t.Run("create fail", func(t *testing.T) {
		repo, mockDB := DomainUserRepoWithSqlMock()
		user, err := NewUser(entity.EmptyUser{})
		if err != nil {
			t.Fatal(err)
		}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err = repo.create(ctx, mockDB.DB, user)
		assert.Equal(t, puddle.ErrClosedPool, err)
	})
}

func TestDomainUserRepo_upsert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainUserRepoWithSqlMock()
		user, err := NewUser(entity.EmptyUser{})
		if err != nil {
			t.Fatal(err)
		}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		cmdTag := pgconn.CommandTag(`1`)
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec", args...).Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err = repo.UpsertMultiple(ctx, mockDB.DB, true, user)
		assert.Nil(t, err)
	})
	t.Run("create fail", func(t *testing.T) {
		repo, mockDB := DomainUserRepoWithSqlMock()
		user, err := NewUser(entity.EmptyUser{})
		if err != nil {
			t.Fatal(err)
		}

		_, userValues := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec", args...).Return(nil, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err = repo.UpsertMultiple(ctx, mockDB.DB, true, user)
		assert.Equal(t, InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec")}.Error(), err.Error())
	})
}

func TestDomainUserRepo_GetByUserNames(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	usernames := []string{"usernames01", "usernames02"}
	_, userRepoEnt := new(User).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainUserRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  usernames,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			req:         usernames,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			req:         usernames,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByUserNames(ctx, mockDB.DB, tt.req.([]string))
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDomainUserRepo_GetByEmails(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	emails := []string{"test-01@test.com", "test-02@test.com"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainUserRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  emails,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			req:         emails,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			req:         emails,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByEmails(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestDomainUserRepo_GetByEmailsInsensitiveCase(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	emails := []string{"test-01@test.com", "TEST-02@test.com"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainUserRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  emails,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "db Query returns error",
			req:  emails,
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name: "rows Scan returns error",
			req:  emails,
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByEmailsInsensitiveCase(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestDomainUserRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userIDs := []string{"user-id-001", "user-id-002"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainUserRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  userIDs,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "db Query returns error",
			req:  userIDs,
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name: "rows Scan returns error",
			req:  userIDs,
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByIDs(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}

}

func TestDomainUserRepo_GetByExternalUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	externalUserIDs := []string{"test-01@test.com", "test-02@test.com"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainUserRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  externalUserIDs,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "db Query returns error",
			req:  externalUserIDs,
			expectedErr: InternalError{
				errors.Wrap(pgx.ErrTxClosed, "db.Query"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name: "rows Scan returns error",
			req:  externalUserIDs,
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByExternalUserIDs(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestGenerateUpdateUserPlaceholders(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		expected := "name = EXCLUDED.name"

		actual := generateUpdateUserPlaceholders([]string{"name"}, true)

		assert.Equal(t, actual, expected)
	})

	t.Run("happy case with ignored fields", func(t *testing.T) {
		expected := "name = EXCLUDED.name"

		actual := generateUpdateUserPlaceholders([]string{"name", "created_at", "deleted_at", "device_token", "avatar", "allow_notification", "last_login_date", "is_tester", "facebook_id", "phone_verified", "email_verified"}, true)

		assert.Equal(t, actual, expected)
	})
}
