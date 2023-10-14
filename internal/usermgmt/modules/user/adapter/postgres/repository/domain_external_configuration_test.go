package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainExternalConfigurationWithSqlMock() (*DomainExternalConfigurationRepo, *testutil.MockDB) {
	r := &DomainExternalConfigurationRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainExternalConfigurationRepo_GetConfigurationByKeys(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	configKeys := []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}
	repo, mockDB := DomainExternalConfigurationWithSqlMock()
	_, domainExternalConfigurationValues := (&ExternalConfiguration{}).FieldMap()
	argsDomainExternalConfiguration := append([]interface{}{}, genSliceMock(len(domainExternalConfigurationValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(configKeys)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainExternalConfiguration...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(configKeys)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			expectedErr: fmt.Errorf("row.Scan: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(configKeys)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainExternalConfiguration...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Err").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetConfigurationByKeys(ctx, mockDB.DB, configKeys)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}
