package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainLocationRepoWithSqlMock() (*DomainLocationRepo, *testutil.MockDB) {
	r := &DomainLocationRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainLocationRepo_GetByPartnerInternalIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainLocationRepoWithSqlMock()
	partnerIDs := []string{uuid.NewString()}
	_, domainLocationValues := NewLocation(entity.NullDomainLocation{}).FieldMap()
	argsDomainLocations := append([]interface{}{}, genSliceMock(len(domainLocationValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(partnerIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainLocations...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(partnerIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			expectedErr: fmt.Errorf("row.Scan: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(partnerIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainLocations...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByPartnerInternalIDs(ctx, mockDB.DB, partnerIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestDomainLocationRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainLocationRepoWithSqlMock()
	locationIDs := []string{uuid.NewString()}
	_, domainLocationValues := NewLocation(entity.NullDomainLocation{}).FieldMap()
	argsDomainLocations := append([]interface{}{}, genSliceMock(len(domainLocationValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(locationIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainLocations...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(locationIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			expectedErr: fmt.Errorf("row.Scan: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(locationIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainLocations...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByIDs(ctx, mockDB.DB, locationIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestDomainLocationRepo_RetrieveLowestLevelLocations(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := DomainLocationRepoWithSqlMock()
	_, domainLocationValues := NewLocation(entity.NullDomainLocation{}).FieldMap()
	argsDomainLocations := append([]interface{}{}, genSliceMock(len(domainLocationValues))...)
	testCases := []struct {
		name        string
		inputName   string
		inputLimit  int32
		inputOffset int32
		inputIDs    []string
		wantError   error
		setup       func()
	}{
		{
			name:        "valid input with no IDs",
			inputName:   "New York",
			inputLimit:  10,
			inputOffset: 0,
			inputIDs:    nil,
			wantError:   nil,
			setup: func() {

				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Once().Return(mockDB.Rows, nil)

				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainLocations...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Next").Once().Return(false)

				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "valid input with IDs",
			inputName:   "New York",
			inputLimit:  0,
			inputOffset: 0,
			inputIDs:    []string{"NYC", "NY"},
			wantError:   nil,
			setup: func() {

				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
					Once().Return(mockDB.Rows, nil)

				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainLocations...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Next").Once().Return(false)

				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Err").Once().Return(nil)
			},
		},
	}

	// Loop over the test cases
	for _, tc := range testCases {
		// Use t.Run to run each subtest
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			repo.RetrieveLowestLevelLocations(ctx, mockDB.DB, tc.inputName, tc.inputLimit, tc.inputOffset, tc.inputIDs)
		})
	}
}
