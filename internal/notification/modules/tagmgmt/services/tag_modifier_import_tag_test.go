package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ImportTags(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	tagRepo := &mock_repositories.MockTagRepo{}
	svc := &TagMgmtModifierService{
		DB:      mockDB,
		TagRepo: tagRepo,
	}
	testCases := []struct {
		Name    string
		Request interface{}
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Request: &npb.ImportTagsRequest{
				Payload: []byte(`tag_id,tag_name,is_archived
				tag-id-1,tag name 1,0`),
			},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				tagRepo.On("FindTagIDsNotExist", ctx, mock.Anything, database.TextArray([]string{"tag-id-1"})).Once().Return(nil, nil)
				tagRepo.On("FindDuplicateTagNames", ctx, mock.Anything, mock.Anything).Once().Return(make(map[string]string), nil)
				tagRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := svc.ImportTags(ctx, testCase.Request.(*npb.ImportTagsRequest))
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
