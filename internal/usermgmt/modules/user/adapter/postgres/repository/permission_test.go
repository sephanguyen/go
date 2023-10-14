package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPermissionRepo_CreateBatch(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	permissionRepo := &PermissionRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.Permission{
				{PermissionID: pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present}},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple permissions",
			req: []*entity.Permission{
				{PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present}},
				{PermissionID: pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present}},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.Permission{
				{PermissionID: pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present}},
			},
			expectedErr: fmt.Errorf("batchResults.Exec %s", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "send batch return not inserted",
			req: []*entity.Permission{
				{PermissionID: pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present}},
			},
			expectedErr: fmt.Errorf("permission not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(failedTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := permissionRepo.CreateBatch(ctx, db, testCase.req.([]*entity.Permission))
		assert.Equal(t, err, testCase.expectedErr)
	}
}
