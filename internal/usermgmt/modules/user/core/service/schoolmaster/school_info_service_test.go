package schoolmaster

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestSchoolInfoModifierService_ImportSchoolInfo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	schoolInfoRepo := new(mock_repositories.MockSchoolInfoRepo)

	// init data
	ids := []interface{}{idutil.ULIDNow(), idutil.ULIDNow()}
	for _, id := range ids {
		schoolInfoRepo.On("Create", ctx, db, mock.Anything).Once().Return(nil)
		schoolInfoRepo.Create(ctx, db, &entity.SchoolInfo{
			ID: database.Text(id.(string)),
		})
	}

	jsm := new(mock_nats.JetStreamManagement)

	s := &SchoolInfoService{
		DB:             db,
		SchoolInfoRepo: schoolInfoRepo,
		JSM:            jsm,
	}

	testcases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			expectedResp: &pb.ImportSchoolInfoResponse{
				Errors: []*pb.ImportSchoolInfoResponse_ImportSchoolInfoError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse school_info item: %s", fmt.Errorf("error parsing IsArchived")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf("unable to parse school_info item: %s", fmt.Errorf("missing mandatory column")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to create school_info item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf("unable to update school_info item: %s", pgx.ErrTxClosed),
					},
				},
			},
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(fmt.Sprintf(`school_id,school_name,school_name_phonetic,school_level_id,address,is_archived
				,School 1,S1,school_level_id-1,Address 1,0
				,School 2,S2,school_level_id-2,Address 2,random-text
				,School 3,S3,school_level_id-3,Address 3,
				,School 4,S4,school_level_id-4,Address 4,1
				%s,School 5,S5,school_level_id-5,Address 5,true
				%s,School 6,S6,school_level_id-6,Address 6,false`, ids...)),
			},
			setup: func(ctx context.Context) {
				schoolInfoRepo.On("BulkImport", ctx, db, mock.Anything).Once().Return([]*repository.ImportError{})
			},
		},
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &pb.ImportSchoolInfoRequest{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - mismatched number of fields in header and content",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, ""),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,school_name
				,1,School 1
				,2,School 2
				,3,School 3`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - number of column != 6",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 6"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,school_name
				1,School 1
				2,School 2
				3,School 3`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != school_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'school_id'"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`School ID,school_name,school_name_phonetic,school_level_id,address,is_archived
				,School 1,S1,school_level_id-1,Address 1,0`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != school_name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'school_name'"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,School Name,school_name_phonetic,school_level_id,address,is_archived
				,School 1,S1,school_level_id-1,Address 1,0`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != school_name_phonetic",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'school_name_phonetic'"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,school_name,School Name Phonetic,school_level_id,address,is_archived
				,School 1,S1,school_level_id-1,Address 1,0`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - fourth column name (toLowerCase) != school_level_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'school_level_id'"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,school_name,school_name_phonetic,School level id,address,is_archived
				,School 1,S1,school_level_id-1,Address 1,0`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - fifth column name (toLowerCase) != address",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'address'"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,school_name,school_name_phonetic,school_level_id,Addresses,is_archived
				,School 1,S1,school_level_id-1,Address 1,0`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - sixth column name (toLowerCase) != is_archived",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - sixth column (toLowerCase) should be 'is_archived'"),
			req: &pb.ImportSchoolInfoRequest{
				Payload: []byte(`school_id,school_name,school_name_phonetic,school_level_id,address,Is Archived
				,School 1,S1,school_level_id-1,Address 1,0`),
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.ImportSchoolInfo(testCase.ctx, testCase.req.(*pb.ImportSchoolInfoRequest))
			if err != nil {
				fmt.Println(err)
			}
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*pb.ImportSchoolInfoResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, schoolInfoRepo)
		})
	}
}
