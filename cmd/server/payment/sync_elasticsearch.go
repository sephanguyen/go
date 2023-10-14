package payment

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/payment/configurations"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/search"
	elasticSearchService "github.com/manabie-com/backend/internal/payment/services/domain_service/elastic_search"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	"github.com/manabie-com/backend/internal/payment/utils"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

var (
	// flags for payment_sync_elasticsearch job
	renewESIndex bool
	schoolID     string
	schoolName   string
)

func init() {
	bootstrap.RegisterJob("payment_sync_elasticsearch", runSyncElasticsearch).
		Desc("Start sync data to elasticsearch").
		BoolVar(&renewESIndex, "renewESIndex", false, "sync to ES without clear index").
		StringVar(&schoolID, "schoolID", "", "sync for specific school").
		StringVar(&schoolName, "schoolName", "", "should match with school name in secret config, for sanity check")
}

func runSyncElasticsearch(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	// func RunSyncElasticsearch(ctx context.Context, c *configurations.Config, shouldRenewESIndex bool, schoolID, schoolName string) {
	zapLogger := rsc.Logger()

	// for db RLS query
	ctx = auth.InjectFakeJwtToken(ctx, schoolID)

	db := rsc.DBWith("fatima")

	searchClient, err := elastic.NewSearchFactory(zapLogger, c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password, "", "")
	if err != nil {
		return fmt.Errorf("unable to connect elasticsearch: %s", err)
	}

	elasticSearch := search.NewElasticSearch(searchClient)
	order := &orderService.OrderService{}

	if err := syncESDocuments(ctx, db, zapLogger, elasticSearch, order, renewESIndex); err != nil {
		return fmt.Errorf("sync data into Elasticsearch failed: %s", err)
	}
	return nil
}

func syncESDocuments(ctx context.Context, db database.QueryExecer, zapLogger *zap.Logger, elasticSearch search.Engine, orderService *orderService.OrderService, shouldRenewESIndex bool) error {
	isUpdatedIndex := syncESDocument(
		constant.ElasticOrderTableName,
		constant.ElasticOrderIndexMapping,
		zapLogger,
		elasticSearch,
		shouldRenewESIndex,
	)
	isUpdatedIndex = isUpdatedIndex || syncESDocument(
		constant.ElasticOrderItemTableName,
		constant.ElasticOrderItemIndexMapping,
		zapLogger,
		elasticSearch,
		shouldRenewESIndex,
	)
	isUpdatedIndex = isUpdatedIndex || syncESDocument(
		constant.ElasticProductTableName,
		constant.ElasticProductIndexMapping,
		zapLogger,
		elasticSearch,
		shouldRenewESIndex,
	)
	if !isUpdatedIndex {
		zapLogger.Info("nothing to update")
		return nil
	} else {
		zapLogger.Info("has new index, need to be synced")
	}

	ordersSync, err := orderService.GetAllOrdersFromDB(ctx, db)
	if err != nil {
		return err
	}
	elasticSearchClient := elasticSearchService.NewElasticSearchService(elasticSearch)
	for _, orderSync := range ordersSync {
		orderItems := make([]entities.OrderItem, 0, len(orderSync.OrderItems))
		mapProducts := make(map[string]bool)
		products := make([]entities.Product, 0)
		for _, orderItem := range orderSync.OrderItems {
			product := orderItem.Product
			orderItemID := pgtype.Text{String: fmt.Sprintf("%s-%s", orderSync.ID.String, product.ID.String)}
			orderItems = append(orderItems, entities.OrderItem{
				OrderID:      orderSync.ID,
				ProductID:    product.ID,
				OrderItemID:  orderItemID,
				DiscountID:   orderItem.DiscountID,
				StartDate:    orderItem.StartDate,
				CreatedAt:    orderItem.CreatedAt,
				ResourcePath: orderItem.ResourcePath,
			})
			ok := mapProducts[product.ID.String]
			if ok {
				continue
			}
			products = append(products, entities.Product{
				ProductID:            product.ID,
				Name:                 product.Name,
				ProductType:          product.ProductType,
				TaxID:                product.TaxID,
				AvailableFrom:        product.AvailableFrom,
				AvailableUntil:       product.AvailableUntil,
				CustomBillingPeriod:  product.CustomBillingPeriod,
				BillingScheduleID:    product.BillingScheduleID,
				DisableProRatingFlag: product.DisableProRatingFlag,
				Remarks:              product.Remarks,
				IsArchived:           product.IsArchived,
				UpdatedAt:            product.UpdatedAt,
				CreatedAt:            product.CreatedAt,
				ResourcePath:         product.ResourcePath,
			})
		}
		err = elasticSearchClient.InsertOrderData(ctx, utils.ElasticSearchData{
			Order: entities.Order{
				OrderID:             orderSync.ID,
				StudentID:           orderSync.StudentID,
				StudentFullName:     orderSync.StudentFullName,
				LocationID:          orderSync.LocationID,
				OrderSequenceNumber: orderSync.OrderSequenceNumber,
				OrderComment:        orderSync.OrderComment,
				OrderStatus:         orderSync.OrderStatus,
				OrderType:           orderSync.OrderType,
				UpdatedAt:           orderSync.UpdatedAt,
				CreatedAt:           orderSync.CreatedAt,
				ResourcePath:        orderSync.ResourcePath,
				IsReviewed:          orderSync.IsReviewed,
			},
			OrderItems: orderItems,
			Products:   products,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func syncESDocument(indexName, index string, zapLogger *zap.Logger, elasticSearch search.Engine, shouldRenewESIndex bool) bool {
	zapLogger.Info(fmt.Sprintf(`creating index "%s" now`, indexName))

	isExisted, err := elasticSearch.CheckIndexExists(indexName)
	if err != nil {
		zapLogger.Fatal("check index exists failed", zap.Error(err))
		return false
	}
	if isExisted {
		if !shouldRenewESIndex {
			return true
		}
		// Delete old index
		respDeleteIndex, err := elasticSearch.DeleteIndex(indexName)
		if err != nil {
			zapLogger.Fatal("delete index failed", zap.Error(err))
			return false
		}
		defer respDeleteIndex.Body.Close()
		if respDeleteIndex.StatusCode != http.StatusOK {
			zapLogger.Fatal(fmt.Sprintf(`unable to delete "%s" index: %s`, indexName, respDeleteIndex.String()))
			return false
		}
		zapLogger.Info(fmt.Sprintf(`"%s" index deleted!`, indexName))
	}

	// Create new index
	idxMap := strings.NewReader(index)
	respCreateIndex, err := elasticSearch.CreateIndex(indexName, idxMap)
	if err != nil {
		zapLogger.Fatal("create index failed", zap.Error(err))
		return false
	}
	defer respCreateIndex.Body.Close()
	if respCreateIndex.StatusCode != http.StatusOK {
		zapLogger.Fatal(fmt.Sprintf(`unable to create "%s" index: %s`, indexName, respCreateIndex.String()))
		return false
	}
	zapLogger.Info(fmt.Sprintf(`"%s" index created!`, indexName))
	return true
}
