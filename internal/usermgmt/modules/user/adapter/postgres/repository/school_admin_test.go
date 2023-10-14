package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSchoolAdminRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	schoolAdminRepo := &SchoolAdminRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.SchoolAdmin{
				{
					SchoolAdminID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			name: "happy case: multiple entities",
			req: []*entity.SchoolAdmin{
				{
					SchoolAdminID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					SchoolAdminID: pgtype.Text{String: "2", Status: pgtype.Present},
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
			req: []*entity.SchoolAdmin{
				{
					SchoolAdminID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "send batch return ",
			req: []*entity.SchoolAdmin{
				{
					SchoolAdminID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("schoolAdmin not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := schoolAdminRepo.CreateMultiple(ctx, db, testCase.req.([]*entity.SchoolAdmin))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestSchoolAdminRepo_Upsert(t *testing.T) {
	t.Parallel()
	schoolAdminRepo := &SchoolAdminRepo{}
	db := &mock_database.Ext{}
	uid := idutil.ULIDNow()
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entity.SchoolAdmin{
				SchoolAdminID: database.Text(uid),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}

				cmdTag := pgconn.CommandTag(successTag)
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: &entity.SchoolAdmin{
				SchoolAdminID: database.Text(uid),
			},
			expectedErr: puddle.ErrClosedPool,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name: "no rows affected",
			req: &entity.SchoolAdmin{
				SchoolAdminID: database.Text(uid),
			},
			expectedErr: fmt.Errorf("cannot upsert schoolAdmin %s", uid),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag(failedTag)
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := schoolAdminRepo.Upsert(ctx, db, testCase.req.(*entity.SchoolAdmin))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestSchoolAdminRepo_SoftDelete(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	schoolAdminRepo := &SchoolAdminRepo{}
	returnAmount := 1

	testCases := []TestCase{
		{
			name:        "error cannot delete schoolAdmin",
			expectedErr: fmt.Errorf("error"),
			setup: func(ctx context.Context) {
				mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(fmt.Sprint(returnAmount)))
				mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				mockDB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := schoolAdminRepo.SoftDelete(ctx, mockDB, database.Text(idutil.ULIDNow()))
		assert.Equal(t, err, testCase.expectedErr)
	}
}

func TestSchoolAdminRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	schoolAdminRepo := &SchoolAdminRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.SchoolAdmin{
				{SchoolAdminID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(successTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: upsert multiple entities",
			req: []*entity.SchoolAdmin{
				{SchoolAdminID: database.Text(idutil.ULIDNow())},
				{SchoolAdminID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(successTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.SchoolAdmin{
				{SchoolAdminID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %v", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "school admin is not upserted",
			req: []*entity.SchoolAdmin{
				{SchoolAdminID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: fmt.Errorf("school admin is not upserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(failedTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := schoolAdminRepo.UpsertMultiple(ctx, db, testCase.req.([]*entity.SchoolAdmin))
		assert.Equal(t, testCase.expectedErr, err)
	}
}
