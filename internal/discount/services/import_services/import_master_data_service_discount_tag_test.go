package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportDiscountTag(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	discountTagRepo := new(mockRepositories.MockDiscountTagRepo)

	s := &ImportMasterDataService{
		DB:              db,
		DiscountTagRepo: discountTagRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportDiscountTagRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportDiscountTagRequest{
				Payload: []byte(`discount_tag_id,discount_tag_name,selectable,is_archived
				,tag_1,true,true,
				,tag_2,true,true,
				,tag_3,true,true,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 4",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 4"),
			Req: &pb.ImportDiscountTagRequest{
				Payload: []byte(`discount_tag_name,selectable,is_archived
				tag_1,true,false
				tag_2,true,false
				tag_3,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != discount_tag_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'discount_tag_id'"),
			Req: &pb.ImportDiscountTagRequest{
				Payload: []byte(`Number,discount_tag_name,selectable,is_archived
				,tag_1,true,false
				,tag_2,true,false
				,tag_3,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != discount_tag_name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'discount_tag_name'"),
			Req: &pb.ImportDiscountTagRequest{
				Payload: []byte(`discount_tag_id,Number,selectable,is_archived
				,tag_1,true,false
				,tag_2,true,false
				,tag_3,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != selectable",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'selectable'"),
			Req: &pb.ImportDiscountTagRequest{
				Payload: []byte(`discount_tag_id,discount_tag_name,Number,is_archived
				,tag_1,true,false
				,tag_2,true,false
				,tag_3,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportDiscountTagRequest{
				Payload: []byte(`discount_tag_id,discount_tag_name,selectable,is_archived
				,tag_1,true,false
				,tag_2,2,false
				,tag_3,,false
				01GXDQSNMPQCJG53EPZ3VQ123R,tag_4,true,false
				,tag_5,true,false
				01GXDQSNMPQC12345PZ3VQCNPR,tag_6,true,false`),
			},
			ExpectedResp: &pb.ImportDiscountTagResponse{
				Errors: []*pb.ImportDiscountTagResponse_ImportDiscountTagError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse discount tag item: error parsing selectable: strconv.ParseBool: parsing \"2\": invalid syntax"),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf("unable to parse discount tag item: missing mandatory data: selectable"),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf("unable to create new discount tag item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf("unable to update discount tag item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				discountTagRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				discountTagRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				discountTagRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				discountTagRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			resp, err := s.ImportDiscountTag(testCase.Ctx, testCase.Req.(*pb.ImportDiscountTagRequest))
			if err != nil {
				fmt.Println(err)
			}
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportDiscountTagResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
		})
	}
}
