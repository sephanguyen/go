package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainSchoolRepoWithSqlMock() (*DomainSchoolRepo, *testutil.MockDB) {
	r := &DomainSchoolRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainSchoolRepo_GetByPartnerInternalIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	partnerIDs := []string{uuid.NewString()}
	repo, mockDB := DomainSchoolRepoWithSqlMock()
	_, domainSchoolValues := NewSchool(entity.DefaultDomainSchool{}).FieldMap()
	argsDomainSchools := append([]interface{}{}, genSliceMock(len(domainSchoolValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(partnerIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainSchools...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "db Query returns error",
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(partnerIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name: "rows Scan returns error",
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(partnerIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainSchools...).Once().Return(pgx.ErrTxClosed)
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

func TestDomainSchoolRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	schoolIDs := []string{uuid.NewString()}
	repo, mockDB := DomainSchoolRepoWithSqlMock()
	_, domainSchoolValues := NewSchool(entity.DefaultDomainSchool{}).FieldMap()
	argsDomainSchools := append([]interface{}{}, genSliceMock(len(domainSchoolValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(schoolIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainSchools...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(schoolIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(schoolIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainSchools...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByIDs(ctx, mockDB.DB, schoolIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestDomainSchoolRepo_GetByGradeIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	schoolIDs := []string{uuid.NewString()}
	gradeID := uuid.NewString()
	repo, mockDB := DomainSchoolRepoWithSqlMock()
	_, domainSchoolValues := NewSchool(entity.DefaultDomainSchool{}).FieldMap()
	argsDomainSchools := append([]interface{}{}, genSliceMock(len(domainSchoolValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(gradeID), database.TextArray(schoolIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainSchools...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(gradeID), database.TextArray(schoolIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(gradeID), database.TextArray(schoolIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainSchools...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByIDsAndGradeID(ctx, mockDB.DB, schoolIDs, gradeID)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}
