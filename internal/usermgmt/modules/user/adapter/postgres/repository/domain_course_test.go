package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func DomainCourseRepoWithSqlMock() (*DomainCourseRepo, *testutil.MockDB) {
	r := &DomainCourseRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainCourseRepo_GetByCoursePartnerIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := DomainCourseRepoWithSqlMock()
	coursePartnerIDs := []string{uuid.NewString()}
	_, domainCourseValues := NewCourse(entity.DefaultDomainCourse{}).FieldMap()
	argsDomainCourses := append([]interface{}{}, genSliceMock(len(domainCourseValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(coursePartnerIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainCourses...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(coursePartnerIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(coursePartnerIDs)).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainCourses...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByCoursePartnerIDs(ctx, mockDB.DB, coursePartnerIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}
