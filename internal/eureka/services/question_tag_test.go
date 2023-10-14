package services

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_validateCSVFormatQuestionTag(t *testing.T) {
	sc := scanner.NewCSVScanner(bytes.NewReader([]byte("id,name,question_tag_type_id")))
	scMissColumn := scanner.NewCSVScanner(bytes.NewReader([]byte("id,name")))
	sc1 := scanner.NewCSVScanner(bytes.NewReader([]byte("wrong_id,name,question_tag_type_id")))
	sc2 := scanner.NewCSVScanner(bytes.NewReader([]byte("id,wrong_name,question_tag_type_id")))
	sc3 := scanner.NewCSVScanner(bytes.NewReader([]byte("id,name,wrong_question_tag_type_id")))

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         sc,
			expectedErr: nil,
		},
		{
			name:        "empty payload",
			req:         scanner.NewCSVScanner(bytes.NewReader([]byte(""))),
			expectedErr: fmt.Errorf("no data in csv file"),
		},
		{
			name:        "missing column",
			req:         scMissColumn,
			expectedErr: fmt.Errorf("csv file invalid format - number of column should be 3"),
		},
		{
			name:        "wrong first column",
			req:         sc1,
			expectedErr: fmt.Errorf("csv file invalid format - first column (toLowerCase) should be 'id'"),
		},
		{
			name:        "wrong second column",
			req:         sc2,
			expectedErr: fmt.Errorf("csv file invalid format - second column (toLowerCase) should be 'name'"),
		},
		{
			name:        "wrong third column",
			req:         sc3,
			expectedErr: fmt.Errorf("csv file invalid format - third column (toLowerCase) should be 'question_tag_type_id'"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateCSVFormatQuestionTag(testCase.req.(scanner.CSVScanner))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestQuestionTagRepo_ImportQuestionTag(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockQuestionTagRepo := &mock_repositories.MockQuestionTagRepo{}

	service := QuestionTagService{
		DB:              mockDB,
		QuestionTagRepo: mockQuestionTagRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockQuestionTagRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			req: &sspb.ImportQuestionTagRequest{
				Payload: []byte(
					`id,name,question_tag_type_id
					id-1,name-1,question-tag-type-id-1
					id-2,name-2,question-tag-type-id-2
					`),
			},
			expectedErr: nil,
		},
		{
			name: "happy case with empty id",
			setup: func(ctx context.Context) {
				mockQuestionTagRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			req: &sspb.ImportQuestionTagRequest{
				Payload: []byte(
					`id,name,question_tag_type_id
					,name-1,question-tag-type-id-1
					,name-2,question-tag-type-id-2
					`),
			},
			expectedErr: nil,
		},
		{
			name:  "name empty",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagRequest{
				Payload: []byte(
					`id,name,question_tag_type_id
					id-1,,question-tag-type-id-1
					`),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprint("s.ImportQuestionTag cannot convert to question tag entity, err: name cannot be empty, at line 2")),
		},
		{
			name:  "question tag type id empty",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagRequest{
				Payload: []byte(
					`id,name,question_tag_type_id
					id-1,name-1,
					`),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprint("s.ImportQuestionTag cannot convert to question tag entity, err: question tag type id cannot be empty, at line 2")),
		},
		{
			name:  "empty csv file",
			setup: func(ctx context.Context) {},
			req: &sspb.ImportQuestionTagRequest{
				Payload: []byte("id,name,question_tag_type_id"),
			},
			expectedErr: status.Error(codes.InvalidArgument, "s.ImportQuestionTag, err: no data in csv file"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := service.ImportQuestionTag(ctx, testCase.req.(*sspb.ImportQuestionTagRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.IsType(t, testCase.expectedResp, resp)
			}
		})
	}
}

func Test_validateCheckDuplicatedQuestionTagIDs(t *testing.T) {
	questionTag1 := []*entities.QuestionTag{
		{
			QuestionTagID: database.Text("question-tag-id-1"),
		},
		{
			QuestionTagID: database.Text("question-tag-id-2"),
		},
		{
			QuestionTagID: database.Text("question-tag-id-3"),
		},
	}
	questionTag2 := []*entities.QuestionTag{
		{
			QuestionTagID: database.Text("question-tag-id-1"),
		},
		{
			QuestionTagID: database.Text("question-tag-id-1"),
		},
		{
			QuestionTagID: database.Text("question-tag-id-3"),
		},
	}
	questionTag3 := []*entities.QuestionTag{
		{
			QuestionTagID: database.Text("question-tag-id-1"),
		},
		{
			QuestionTagID: database.Text("question-tag-id-3"),
		},
		{
			QuestionTagID: database.Text("question-tag-id-3"),
		},
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         questionTag1,
			expectedErr: nil,
		},
		{
			name:        "duplicate id",
			req:         questionTag2,
			expectedErr: fmt.Errorf("duplicated id: question-tag-id-1 at line 2 and 3"),
		},
		{
			name:        "duplicate id 3 and 4",
			req:         questionTag3,
			expectedErr: fmt.Errorf("duplicated id: question-tag-id-3 at line 3 and 4"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := checkDuplicatedQuestionTagIDs(testCase.req.([]*entities.QuestionTag))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
