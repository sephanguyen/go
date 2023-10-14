package services

import (
	"context"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductGroup_RetrieveProductGroupsOfProductIDByDiscountType(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                      *mockDb.Ext
		productGroupRepo        *mockRepositories.MockProductGroupRepo
		productGroupMappingRepo *mockRepositories.MockProductGroupMappingRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productGroupMappingRepo.On("GetByProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.ProductGroupMapping{
					{
						ProductGroupID: pgtype.Text{String: mock.Anything},
						ProductID:      pgtype.Text{String: mock.Anything},
					},
				}, nil)
				productGroupRepo.On("GetByID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductGroup{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when getting product group mapping",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productGroupMappingRepo.On("GetByProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.ProductGroupMapping{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when geting product group",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productGroupMappingRepo.On("GetByProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.ProductGroupMapping{
					{
						ProductGroupID: pgtype.Text{String: mock.Anything},
						ProductID:      pgtype.Text{String: mock.Anything},
					},
				}, nil)
				productGroupRepo.On("GetByID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductGroup{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productGroupRepo = new(mockRepositories.MockProductGroupRepo)
			productGroupMappingRepo = new(mockRepositories.MockProductGroupMappingRepo)

			testCase.Setup(testCase.Ctx)
			s := &ProductGroupService{
				DB:                      db,
				ProductGroupRepo:        productGroupRepo,
				ProductGroupMappingRepo: productGroupMappingRepo,
			}
			_, err := s.RetrieveProductGroupsOfProductIDByDiscountType(testCase.Ctx, mock.Anything, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productGroupRepo, productGroupMappingRepo)
		})
	}
}

func TestProductGroup_RetrieveEligibleProductGroupsOfStudentProductsByDiscountType(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                      *mockDb.Ext
		productGroupRepo        *mockRepositories.MockProductGroupRepo
		productGroupMappingRepo *mockRepositories.MockProductGroupMappingRepo
	)

	studentProducts := []entities.StudentProduct{{
		ProductID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
	}}

	testcases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedResp: []entities.ProductDiscountGroup{
				{
					StudentProduct: studentProducts[0],
					ProductGroups: []entities.ProductGroup{{
						ProductGroupID: pgtype.Text{String: mock.Anything},
						DiscountType:   pgtype.Text{String: mock.Anything},
					}},
					DiscountType: mock.Anything,
				},
			},
			Setup: func(ctx context.Context) {
				productGroupMappingRepo.On("GetByProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.ProductGroupMapping{
					{
						ProductGroupID: pgtype.Text{String: mock.Anything},
						ProductID:      pgtype.Text{String: mock.Anything},
					},
				}, nil)
				productGroupRepo.On("GetByID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductGroup{
					ProductGroupID: pgtype.Text{String: mock.Anything},
					DiscountType:   pgtype.Text{String: mock.Anything},
				}, nil)
			},
		},
		{
			Name:         "Fail case: Error when getting product groups",
			Ctx:          interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedResp: []entities.ProductDiscountGroup{},
			Setup: func(ctx context.Context) {
				productGroupMappingRepo.On("GetByProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.ProductGroupMapping{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productGroupRepo = new(mockRepositories.MockProductGroupRepo)
			productGroupMappingRepo = new(mockRepositories.MockProductGroupMappingRepo)

			testCase.Setup(testCase.Ctx)
			s := &ProductGroupService{
				DB:                      db,
				ProductGroupRepo:        productGroupRepo,
				ProductGroupMappingRepo: productGroupMappingRepo,
			}

			resp := s.RetrieveEligibleProductGroupsOfStudentProductsByDiscountType(testCase.Ctx, studentProducts, mock.Anything)
			assert.Equal(t, testCase.ExpectedResp, resp)

			mock.AssertExpectationsForObjects(t, db, productGroupRepo, productGroupMappingRepo)
		})
	}
}
