package services

import (
	"context"
	"fmt"
	"testing"

	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestQuestionTagType_ImportQuestionTagTypes(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	mockQuestiontagTypeRepo := &mock_repositories.MockQuestionTagTypeRepo{}

	service := QuestionTagTypeService{
		DB:                  mockDB,
		QuestionTagTypeRepo: mockQuestiontagTypeRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockQuestiontagTypeRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(
					`id,name
					id-1,name-1
					id-2,name-2
					`),
			},
			expectedErr: nil,
		},
		{
			name: "happy case with empty id",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockQuestiontagTypeRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(
					`id,name
					,name-1
					,name-2
					`),
			},
			expectedErr: nil,
		},
		{
			name:  "empty csv file",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(""),
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateCSVFormat: no data in csv file"),
		},
		{
			name:  "columns not equal to 2",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(`id,name,column3`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateCSVFormat: csv file has invalid format - number of column should be 2"),
		},
		{
			name:  `first column not equal to "id"`,
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(`first,name`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateCSVFormat: csv file has invalid format - first column (toLowerCase) should be 'id'"),
		},
		{
			name:  `second column not equal to "name"`,
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(`id,second`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateCSVFormat: csv file has invalid format - second column (toLowerCase) should be 'name'"),
		},
		{
			name:  "name be empty",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(
					`id,name
					id-1,
					`),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprint("newQuestionTagTypeRow: name be empty! at line 2")),
		},
		{
			name:  "no value rows",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(`id,name`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
		},
		{
			name:  "duplicated ids",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagTypesRequest{
				Payload: []byte(
					`id,name
					id-1,name-1
					id-2,name-2
					id-1,name-2
					`),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("checkDuplicatedQuestionTagTypeIDs: duplicated id: id-1 at line 2 and 4")),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := service.ImportQuestionTagTypes(ctx, testCase.req.(*sspb.ImportQuestionTagTypesRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}
