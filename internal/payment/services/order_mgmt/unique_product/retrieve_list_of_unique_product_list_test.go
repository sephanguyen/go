package unique_product

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRetrieveListOfUniqueProductIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductServiceForUniqueProduct
		packageService        *mockServices.IPackageServiceForUniqueProduct
	)

	TestCases := []utils.TestCase{
		{
			Name: "Error when get student product list of unique product ",
			Ctx:  interceptors.ContextWithUserGroup(ctx, constant.UserGroupSchoolAdmin),
			Req: &pb.RetrieveListOfUniqueProductIDsRequest{
				StudentId: "student_id_1",
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On(
					"GetUniqueProductsByStudentID",
					ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Error when get end date of unique recurring product ",
			Ctx:  interceptors.ContextWithUserGroup(ctx, constant.UserGroupSchoolAdmin),
			Req: &pb.RetrieveListOfUniqueProductIDsRequest{
				StudentId: "student_id_1",
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On(
					"GetUniqueProductsByStudentID",
					ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*entities.StudentProduct{
						{

							ProductID: pgtype.Text{
								String: "product_id_1",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_CANCELLED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_2",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_CANCELLED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_3",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_CANCELLED.String(),
								Status: pgtype.Present,
							},
						},
					}, nil)
				packageService.On("GetByIDForUniqueProduct", ctx, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case of one time product",
			Ctx:  interceptors.ContextWithUserGroup(ctx, constant.UserGroupSchoolAdmin),
			Req: &pb.RetrieveListOfUniqueProductIDsRequest{
				StudentId: "student_id_1",
			},
			ExpectedErr: nil,
			ExpectedResp: &pb.RetrieveListOfUniqueProductIDsResponse{
				ProductDetails: []*pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo{
					{

						ProductId: "product_id_1",
					},
					{
						ProductId: "product_id_2",
					},
					{
						ProductId: "product_id_3",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentProductService.On(
					"GetUniqueProductsByStudentID",
					ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*entities.StudentProduct{
						{

							ProductID: pgtype.Text{
								String: "product_id_1",
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_ORDERED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_2",
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_ORDERED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_3",
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_ORDERED.String(),
								Status: pgtype.Present,
							},
						},
					}, nil)
			},
		},
		{
			Name: "Happy case of recurring product with status not CANCELLED",
			Ctx:  interceptors.ContextWithUserGroup(ctx, constant.UserGroupSchoolAdmin),
			Req: &pb.RetrieveListOfUniqueProductIDsRequest{
				StudentId: "student_id_1",
			},
			ExpectedErr: nil,
			ExpectedResp: &pb.RetrieveListOfUniqueProductIDsResponse{
				ProductDetails: []*pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo{
					{

						ProductId: "product_id_1",
					},
					{
						ProductId: "product_id_2",
					},
					{
						ProductId: "product_id_3",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentProductService.On(
					"GetUniqueProductsByStudentID",
					ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*entities.StudentProduct{
						{

							ProductID: pgtype.Text{
								String: "product_id_1",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_ORDERED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_2",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_ORDERED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_3",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_ORDERED.String(),
								Status: pgtype.Present,
							},
						},
					}, nil)
			},
		},
		{
			Name: "Happy case of recurring product with status CANCELLED",
			Ctx:  interceptors.ContextWithUserGroup(ctx, constant.UserGroupSchoolAdmin),
			Req: &pb.RetrieveListOfUniqueProductIDsRequest{
				StudentId: "student_id_1",
			},
			ExpectedErr: nil,
			ExpectedResp: &pb.RetrieveListOfUniqueProductIDsResponse{
				ProductDetails: []*pb.RetrieveListOfUniqueProductIDsResponse_ProductInfo{
					{

						ProductId: "product_id_1",
						EndTime:   &timestamppb.Timestamp{Seconds: int64(time.Now().Unix())},
					},
					{
						ProductId: "product_id_2",
						EndTime:   &timestamppb.Timestamp{Seconds: int64(time.Now().Unix())},
					},
					{
						ProductId: "product_id_3",
						EndTime:   &timestamppb.Timestamp{Seconds: int64(time.Now().Unix())},
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentProductService.On(
					"GetUniqueProductsByStudentID",
					ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*entities.StudentProduct{
						{

							ProductID: pgtype.Text{
								String: "product_id_1",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_CANCELLED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_2",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_CANCELLED.String(),
								Status: pgtype.Present,
							},
						},
						{

							ProductID: pgtype.Text{
								String: "product_id_3",
								Status: pgtype.Present,
							},
							EndDate: pgtype.Timestamptz{
								Time:   database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
								Status: pgtype.Present,
							},
							ProductStatus: pgtype.Text{
								String: pb.StudentProductStatus_CANCELLED.String(),
								Status: pgtype.Present,
							},
						},
					}, nil)
				packageService.On("GetByIDForUniqueProduct", ctx, mock.Anything, mock.Anything).Return(entities.Package{}, pgx.ErrNoRows)
				studentProductService.On(
					"EndDateOfUniqueRecurringProduct",
					ctx, mock.Anything, mock.Anything, mock.Anything).Return(time.Now(), nil)
			},
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.IStudentProductServiceForUniqueProduct)
			packageService = new(mockServices.IPackageServiceForUniqueProduct)
			s := &UniqueProduct{
				DB:                    db,
				StudentProductService: studentProductService,
				PackageService:        packageService,
			}

			testCase.Setup(testCase.Ctx)

			resp, err := s.RetrieveListOfUniqueProductIDs(testCase.Ctx, testCase.Req.(*pb.RetrieveListOfUniqueProductIDsRequest))
			if err != nil {
				fmt.Println(err)
			}
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductService, packageService)
		})
	}

}
