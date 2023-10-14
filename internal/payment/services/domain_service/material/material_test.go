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

func TestMaterialService_GetMaterialByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		materialRepo *mockRepositories.MockMaterialRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when checking material id: %v", constant.ErrDefault),
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				materialRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Material{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				materialRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Material{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			materialRepo = new(mockRepositories.MockMaterialRepo)

			testCase.Setup(testCase.Ctx)
			s := &MaterialService{
				materialRepo: materialRepo,
			}
			_, err := s.GetMaterialByID(testCase.Ctx, db, testCase.Req.(string))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, materialRepo)
		})
	}
}

func TestMaterialService_GetAllMaterialsForExportService(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		materialRepo *mockRepositories.MockMaterialRepo
		productRepo  *mockRepositories.MockProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get all materials",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				materialRepo.On("GetAll", ctx, db).Return([]entities.Material{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get products by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				materialRepo.On("GetAll", ctx, db).Return([]entities.Material{}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when missing product info",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Missing product info with id"),
			Setup: func(ctx context.Context) {
				materialRepo.On("GetAll", ctx, db).Return([]entities.Material{
					{
						MaterialID: pgtype.Text{
							String: "material_product_1",
							Status: pgtype.Present,
						},
					},
					{
						MaterialID: pgtype.Text{
							String: "material_product_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "material_product_1",
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
				materialRepo.On("GetAll", ctx, db).Return([]entities.Material{
					{
						MaterialID: pgtype.Text{
							String: "material_product_1",
							Status: pgtype.Present,
						},
					},
					{
						MaterialID: pgtype.Text{
							String: "material_product_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "material_product_1",
							Status: pgtype.Present,
						},
					},
					{
						ProductID: pgtype.Text{
							String: "material_product_2",
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
			materialRepo = new(mockRepositories.MockMaterialRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &MaterialService{
				materialRepo: materialRepo,
				productRepo:  productRepo,
			}

			_, err := s.GetAllMaterialsForExport(testCase.Ctx, db)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, materialRepo, productRepo)
		})
	}
}
