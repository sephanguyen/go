package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/search"
	"github.com/manabie-com/backend/internal/payment/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ElasticSearchService struct {
	SearchEngine search.Engine
}

func (s *ElasticSearchService) InsertOrderData(ctx context.Context, data utils.ElasticSearchData) (err error) {
	err = s.insertOrderIntoElastic(ctx, data.Order)
	if err != nil {
		err = status.Errorf(codes.Internal, "inserting order to elastic search have error %v", err.Error())
		return
	}

	err = s.insertProductsIntoElastic(ctx, data.Products)
	if err != nil {
		err = status.Errorf(codes.Internal, "inserting products to elastic search have error %v", err.Error())
		return
	}

	err = s.insertOrderItemsIntoElastic(ctx, data.OrderItems)
	if err != nil {
		err = status.Errorf(codes.Internal, "inserting order items to elastic search have error %v", err.Error())
	}
	return
}

func (s *ElasticSearchService) insertOrderIntoElastic(ctx context.Context, order entities.Order) error {
	_, err := s.SearchEngine.Insert(ctx, constant.ElasticOrderTableName, []search.InsertionContent{
		{
			ID: order.OrderID.String,
			Data: entities.ElasticOrder{
				OrderID:             order.OrderID.String,
				StudentID:           order.StudentID.String,
				StudentName:         order.StudentFullName.String,
				LocationID:          order.LocationID.String,
				OrderSequenceNumber: order.OrderSequenceNumber.Int,
				OrderComment:        order.OrderComment.String,
				OrderStatus:         order.OrderStatus.String,
				OrderType:           order.OrderType.String,
				UpdatedAt:           order.UpdatedAt.Time,
				CreatedAt:           order.CreatedAt.Time,
			},
		},
	})
	return err
}

func (s *ElasticSearchService) insertOrderItemsIntoElastic(ctx context.Context, orderItems []entities.OrderItem) error {
	contents := make([]search.InsertionContent, 0, len(orderItems))
	for _, orderItem := range orderItems {
		contents = append(contents, search.InsertionContent{
			ID: orderItem.OrderItemID.String,
			Data: entities.ElasticOrderItem{
				OrderID:     orderItem.OrderID.String,
				ProductID:   orderItem.ProductID.String,
				OrderItemID: orderItem.OrderItemID.String,
				ProductName: orderItem.ProductName.String,
				DiscountID:  orderItem.DiscountID.String,
				StartDate:   orderItem.StartDate.Time,
				CreatedAt:   orderItem.CreatedAt.Time,
			},
		})
	}
	_, err := s.SearchEngine.Insert(ctx, constant.ElasticOrderItemTableName, contents)
	return err
}

func (s *ElasticSearchService) insertProductsIntoElastic(ctx context.Context, products []entities.Product) error {
	contents := make([]search.InsertionContent, 0, len(products))
	for _, product := range products {
		contents = append(contents, search.InsertionContent{
			ID: fmt.Sprint(product.ProductID.String),
			Data: entities.ElasticProduct{
				ProductID:            product.ProductID.String,
				Name:                 product.Name.String,
				ProductType:          product.ProductType.String,
				TaxID:                product.TaxID.String,
				AvailableFrom:        product.AvailableFrom.Time,
				AvailableUntil:       product.AvailableUntil.Time,
				CustomBillingPeriod:  product.CustomBillingPeriod.Time,
				BillingScheduleID:    product.BillingScheduleID.String,
				DisableProRatingFlag: product.DisableProRatingFlag.Bool,
				Remarks:              product.Remarks.String,
				IsArchived:           product.IsArchived.Bool,
				UpdatedAt:            product.UpdatedAt.Time,
				CreatedAt:            product.CreatedAt.Time,
			},
		})
	}
	_, err := s.SearchEngine.Insert(ctx, constant.ElasticProductTableName, contents)
	return err
}

func NewElasticSearchService(searchEngine search.Engine) *ElasticSearchService {
	return &ElasticSearchService{SearchEngine: searchEngine}
}
