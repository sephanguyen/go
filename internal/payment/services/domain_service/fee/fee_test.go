package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFeeService_GetAllFeesForExportService(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		feeRepo     *mockRepositories.MockFeeRepo
		productRepo *mockRepositories.MockProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get all fees",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				feeRepo.On("GetAll", ctx, db).Return([]entities.Fee{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get products by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				feeRepo.On("GetAll", ctx, db).Return([]entities.Fee{}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when missing product info",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Missing product info with id"),
			Setup: func(ctx context.Context) {
				feeRepo.On("GetAll", ctx, db).Return([]entities.Fee{
					{
						FeeID: pgtype.Text{
							String: "fee_product_id_1",
							Status: pgtype.Present,
						},
					},
					{
						FeeID: pgtype.Text{
							String: "fee_product_id_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "fee_product_id_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				feeRepo.On("GetAll", ctx, db).Return([]entities.Fee{
					{
						FeeID: pgtype.Text{
							String: "fee_product_id_1",
							Status: pgtype.Present,
						},
					},
					{
						FeeID: pgtype.Text{
							String: "fee_product_id_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "fee_product_id_2",
							Status: pgtype.Present,
						},
					},
					{
						ProductID: pgtype.Text{
							String: "fee_product_id_1",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			feeRepo = new(mockRepositories.MockFeeRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &FeeService{
				feeRepo:     feeRepo,
				productRepo: productRepo,
			}

			_, err := s.GetAllFeesForExport(testCase.Ctx, db)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, feeRepo, productRepo)
		})
	}
}
