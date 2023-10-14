package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/search"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestElasticSearchService_insertOrderIntoElastic(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db     *mockDb.Ext
		engine *mockServices.Engine
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when insert order into elastic",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         entities.Order{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticOrderTableName, mock.Anything).Return(1, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         entities.Order{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticOrderTableName, mock.Anything).Return(1, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			engine = new(mockServices.Engine)
			testCase.Setup(testCase.Ctx)
			s := &ElasticSearchService{
				SearchEngine: engine,
			}

			req := testCase.Req.(entities.Order)
			err := s.insertOrderIntoElastic(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, engine)
		})
	}
}

func TestElasticSearchService_insertOrderItemsIntoElastic(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db     *mockDb.Ext
		engine *mockServices.Engine
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when insert order item into elastic",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         []entities.OrderItem{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticOrderItemTableName, mock.Anything).Return(1, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         []entities.OrderItem{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticOrderItemTableName, mock.Anything).Return(1, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			engine = new(mockServices.Engine)
			testCase.Setup(testCase.Ctx)
			s := &ElasticSearchService{
				SearchEngine: engine,
			}

			req := testCase.Req.([]entities.OrderItem)
			err := s.insertOrderItemsIntoElastic(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, engine)
		})
	}
}

func TestElasticSearchService_insertProductsIntoElastic(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db     *mockDb.Ext
		engine *mockServices.Engine
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when insert products into elastic",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         []entities.Product{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticProductTableName, mock.Anything).Return(1, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         []entities.Product{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticProductTableName, mock.Anything).Return(1, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			engine = new(mockServices.Engine)
			testCase.Setup(testCase.Ctx)
			s := &ElasticSearchService{
				SearchEngine: engine,
			}

			req := testCase.Req.([]entities.Product)
			err := s.insertProductsIntoElastic(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, engine)
		})
	}
}

func TestElasticSearchService_InsertOrderData(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db     *mockDb.Ext
		engine *mockServices.Engine
	)
	testcases := []utils.TestCase{
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         utils.ElasticSearchData{},
			Setup: func(ctx context.Context) {
				engine.On("Insert", ctx, constant.ElasticOrderTableName, mock.Anything).Return(1, nil)
				engine.On("Insert", ctx, constant.ElasticProductTableName, mock.Anything).Return(1, nil)
				engine.On("Insert", ctx, constant.ElasticOrderItemTableName, mock.Anything).Return(1, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			engine = new(mockServices.Engine)
			testCase.Setup(testCase.Ctx)
			s := &ElasticSearchService{
				SearchEngine: engine,
			}

			req := testCase.Req.(utils.ElasticSearchData)
			err := s.InsertOrderData(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, engine)
		})
	}
}
