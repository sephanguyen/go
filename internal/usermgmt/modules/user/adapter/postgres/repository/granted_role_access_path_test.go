package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGrantedRoleAccessPath_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	grantedRoleAccessPathRepo := &GrantedRoleAccessPathRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.GrantedRoleAccessPath{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
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
			name: "happy case: upsert multiple granted_role_access_path",
			req: []*entity.GrantedRoleAccessPath{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
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
			name: "success: resource path status null",
			req: []*entity.GrantedRoleAccessPath{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
			},
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
			name: "success: skip location id status not equal present",
			req: []*entity.GrantedRoleAccessPath{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Null},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
			},
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
			name: "error send batch",
			req: []*entity.GrantedRoleAccessPath{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					LocationID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
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
		err := grantedRoleAccessPathRepo.Upsert(ctx, db, testCase.req.([]*entity.GrantedRoleAccessPath))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
