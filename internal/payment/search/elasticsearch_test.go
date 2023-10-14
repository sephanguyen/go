package search

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/search/op"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

const (
	RespOrderFile = "../testbed/search/resp_order.json"
	OrderStatus   = "order-status"
)

type Order struct {
	OrderID             string `json:"order_id"`
	StudentID           string `json:"student-id"`
	LocationID          string `json:"location-id"`
	OrderSequenceNumber int32  `json:"order-sequence-number"`
	OrderComment        string `json:"order-comment"`
	OrderStatus         string `json:"order-status"`
	OrderType           string `json:"order-type"`
	CreatedTime         int64  `json:"created-time"`
}

var testcaseOrders = []Order{
	{
		OrderID:             "a2d3fdf5-7c30-4621-9ea4-fc15b9bdf428",
		StudentID:           "student-id-3",
		LocationID:          "location-id-3",
		OrderSequenceNumber: 14,
		OrderComment:        "order-comment-3",
		OrderStatus:         "5",
		OrderType:           "6",
		CreatedTime:         1646337336,
	},
	{
		OrderID:             "5e684346-06e4-4bc2-a08c-6ff1d48e4afb",
		StudentID:           "student-id-3",
		LocationID:          "location-id-3",
		OrderSequenceNumber: 14,
		OrderComment:        "order-comment-3",
		OrderStatus:         "5",
		OrderType:           "6",
		CreatedTime:         1646337638,
	},
}

func TestGetAll(t *testing.T) {
	mockResp, err := utils.ReadFile(RespOrderFile)
	assert.NoError(t, err)
	mockSearchFactory, closeSearch := elastic.NewMockSearchFactory(string(mockResp))
	defer closeSearch()

	rp := "manabie"
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})

	tableName := "order"
	esSearch := NewElasticSearch(mockSearchFactory)

	orders, err := esSearch.GetAll(ctx, tableName, func(data []byte) (interface{}, error) {
		var order Order
		err = json.Unmarshal(data, &order)
		return order, err
	}, PagingParam{FromIdx: 0, NumberRows: 100}, SortParam{
		ColumnName: "order_id",
		Ascending:  true,
	})
	assert.NoError(t, err)
	assert.Equal(t, len(orders), 2)
	for idx, item := range orders {
		order := item.(Order)
		expectedOrder := testcaseOrders[idx]
		assert.Equal(t, expectedOrder.OrderID, order.OrderID)
		assert.Equal(t, expectedOrder.StudentID, order.StudentID)
		assert.Equal(t, expectedOrder.LocationID, order.LocationID)
		assert.Equal(t, expectedOrder.OrderSequenceNumber, order.OrderSequenceNumber)
		assert.Equal(t, expectedOrder.OrderComment, order.OrderComment)
		assert.Equal(t, expectedOrder.OrderStatus, order.OrderStatus)
		assert.Equal(t, expectedOrder.OrderType, order.OrderType)
		assert.Equal(t, expectedOrder.CreatedTime, order.CreatedTime)
	}
}

func TestSearch(t *testing.T) {
	mockResp, err := utils.ReadFile(RespOrderFile)
	assert.NoError(t, err)
	mockSearchFactory, closeSearch := elastic.NewMockSearchFactory(string(mockResp))
	defer closeSearch()

	rp := "manabie"
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})

	tableName := "order"
	esSearch := NewElasticSearch(mockSearchFactory)

	condition := op.Equal(OrderStatus, strconv.Itoa(int(pb.OrderStatus_ORDER_STATUS_INVOICED)))
	orders, err := esSearch.Search(ctx, tableName, condition, func(data []byte) (interface{}, error) {
		var order Order
		err = json.Unmarshal(data, &order)
		return order, err
	}, PagingParam{FromIdx: 0, NumberRows: 100}, SortParam{
		ColumnName: "order_id",
		Ascending:  true,
	})
	assert.NoError(t, err)
	assert.Equal(t, len(orders), 2)
	for idx, item := range orders {
		order := item.(Order)
		expectedOrder := testcaseOrders[idx]
		assert.Equal(t, expectedOrder.OrderID, order.OrderID)
		assert.Equal(t, expectedOrder.StudentID, order.StudentID)
		assert.Equal(t, expectedOrder.LocationID, order.LocationID)
		assert.Equal(t, expectedOrder.OrderSequenceNumber, order.OrderSequenceNumber)
		assert.Equal(t, expectedOrder.OrderComment, order.OrderComment)
		assert.Equal(t, expectedOrder.OrderStatus, order.OrderStatus)
		assert.Equal(t, expectedOrder.OrderType, order.OrderType)
		assert.Equal(t, expectedOrder.CreatedTime, order.CreatedTime)
	}
}

func TestSearchWithoutPaging(t *testing.T) {
	mockResp, err := utils.ReadFile(RespOrderFile)
	assert.NoError(t, err)
	mockSearchFactory, closeSearch := elastic.NewMockSearchFactory(string(mockResp))
	defer closeSearch()

	rp := "manabie"
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})

	tableName := "order"
	esSearch := NewElasticSearch(mockSearchFactory)

	condition := op.Equal(OrderStatus, strconv.Itoa(int(pb.OrderStatus_ORDER_STATUS_INVOICED)))
	orders, err := esSearch.SearchWithoutPaging(ctx, tableName, condition, func(data []byte) (interface{}, error) {
		var order Order
		err = json.Unmarshal(data, &order)
		return order, err
	}, SortParam{
		ColumnName: "order_id",
		Ascending:  true,
	})
	assert.NoError(t, err)
	assert.Equal(t, len(orders), 2)
	for idx, item := range orders {
		order := item.(Order)
		expectedOrder := testcaseOrders[idx]
		assert.Equal(t, expectedOrder.OrderID, order.OrderID)
		assert.Equal(t, expectedOrder.StudentID, order.StudentID)
		assert.Equal(t, expectedOrder.LocationID, order.LocationID)
		assert.Equal(t, expectedOrder.OrderSequenceNumber, order.OrderSequenceNumber)
		assert.Equal(t, expectedOrder.OrderComment, order.OrderComment)
		assert.Equal(t, expectedOrder.OrderStatus, order.OrderStatus)
		assert.Equal(t, expectedOrder.OrderType, order.OrderType)
		assert.Equal(t, expectedOrder.CreatedTime, order.CreatedTime)
	}
}

func TestCountValue(t *testing.T) {
	mockResp, err := utils.ReadFile("../testbed/search/resp_order_count.json")
	assert.NoError(t, err)
	mockSearchFactory, closeSearch := elastic.NewMockSearchFactory(string(mockResp))
	defer closeSearch()

	rp := "manabie"
	ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})

	tableName := "order"
	esSearch := NewElasticSearch(mockSearchFactory)

	condition := op.Equal(OrderStatus, strconv.Itoa(int(pb.OrderStatus_ORDER_STATUS_INVOICED)))
	value, err := esSearch.CountValue(ctx, tableName, "order_id", condition)
	assert.NoError(t, err)
	assert.Equal(t, value, uint32(2))
}
