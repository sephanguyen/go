package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrganizationRepoWithSqlMock() (*OrganizationRepo, *testutil.MockDB) {
	r := &OrganizationRepo{}
	return r, testutil.NewMockDB()
}

type TestCase struct {
	name         string
	req          interface{}
	expectedErr  error
	expectedResp interface{}
	setup        func(ctx context.Context)
}

func TestOrganizationRepo_GetSalesforceClientIDByOrganizationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := OrganizationRepoWithSqlMock()
	organizationID := uuid.NewString()
	salesforceClientID := ""
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), database.Text(organizationID)).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", &salesforceClientID).Once().Return(nil)

				mockDB.Row.On("Err").Once().Return(nil)
				mockDB.Row.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "row scan returns error",
			expectedErr: repository.InternalError{RawError: pgx.ErrTxClosed},
			setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), database.Text(organizationID)).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", &salesforceClientID).Once().Return(pgx.ErrTxClosed)
				mockDB.Row.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			clientID, err := repo.GetSalesforceClientIDByOrganizationID(ctx, mockDB.DB, organizationID)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, salesforceClientID, clientID)
		})
	}
}
