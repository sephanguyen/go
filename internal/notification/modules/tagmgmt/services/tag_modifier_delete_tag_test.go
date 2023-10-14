package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	mock_noti_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_DeleteTag(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)
	err := pgx.ErrNoRows
	tagRepo := &mock_repositories.MockTagRepo{}
	ifntRepo := &mock_noti_repositories.MockInfoNotificationTagRepo{}
	svc := &TagMgmtModifierService{
		DB:                      mockDB,
		TagRepo:                 tagRepo,
		InfoNotificationTagRepo: ifntRepo,
	}
	tagID := "tag-id-1"
	testCases := []struct {
		Name    string
		Request interface{}
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "handle err tag repo",
			Request: &npb.DeleteTagRequest{TagId: tagID},
			ExpcErr: status.Error(codes.Internal, fmt.Sprintf("DeleteTag.TagRepo.SoftDelete: %v", err)),
			Setup: func(ctx context.Context) {
				tagRepo.On("FindByID", ctx, mock.Anything, database.Text(tagID)).Once().Return(nil, nil)
				ifntRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				tagRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(err)
			},
		},
		{
			Name:    "handle err info noti tag repo",
			Request: &npb.DeleteTagRequest{TagId: tagID},
			ExpcErr: status.Error(codes.Internal, fmt.Sprintf("DeleteTag.InfoNotificationTagRepo.SoftDelete: %v", err)),
			Setup: func(ctx context.Context) {
				tagRepo.On("FindByID", ctx, mock.Anything, database.Text(tagID)).Once().Return(nil, nil)
				ifntRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(err)
				tagRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:    "happy case",
			Request: &npb.DeleteTagRequest{TagId: tagID},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				tagRepo.On("FindByID", ctx, mock.Anything, database.Text(tagID)).Once().Return(nil, nil)
				ifntRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				tagRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:    "empty tag ID",
			Request: &npb.DeleteTagRequest{},
			ExpcErr: status.Error(codes.InvalidArgument, "TagID is empty"),
			Setup:   func(ctx context.Context) {},
		},
		{
			Name:    "tag not exist",
			Request: &npb.DeleteTagRequest{TagId: tagID},
			ExpcErr: status.Error(codes.Internal, fmt.Sprintf("DeleteTag: %v", err)),
			Setup: func(ctx context.Context) {
				tagRepo.On("FindByID", ctx, mock.Anything, database.Text(tagID)).Once().Return(nil, err)
				ifntRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				tagRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := svc.DeleteTag(ctx, testCase.Request.(*npb.DeleteTagRequest))
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
