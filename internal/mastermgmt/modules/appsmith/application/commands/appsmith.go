package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/infrastructure"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppsmithCommandHandler struct {
	DB         *mongo.Database
	LogRepo    infrastructure.LogRepo
	HTTPClient clients.HTTPClientInterface
}

func (a *AppsmithCommandHandler) SaveLog(ctx context.Context, e domain.EventLog) (domain.EventLog, error) {
	dbName := "appsmith"
	conn := "mongodb+srv://root:M%40nabie123@appsmithcluster.g3raoxa.mongodb.net/?retryWrites=true&w=majority"

	// Set client options
	co := options.Client().ApplyURI(conn)
	// Connect to MongoDB
	client, err := mongo.Connect(ctx, co)
	if err != nil {
		return nil, err
	}
	db := client.Database(dbName)
	return a.LogRepo.SaveLog(ctx, db, e)
}

func (a *AppsmithCommandHandler) PullMetadata(ctx context.Context, branchName string, config configs.AppsmithAPI) (*domain.AppsmithResponse, error) {
	endpoint := fmt.Sprintf("%s/git/pull/app/%s", config.ENDPOINT, config.ApplicationID)
	headers := make(clients.Headers)
	headers["Authorization"] = fmt.Sprintf("Basic %s", config.Authorization)
	headers["branchName"] = branchName
	response, err := clients.HandleHTTPRequest[domain.AppsmithResponse](a.HTTPClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodGet,
		URL:     endpoint,
		Body:    nil,
		Headers: &headers,
	})

	return response, err
}

func (a *AppsmithCommandHandler) DiscardChange(ctx context.Context, branchName string, config configs.AppsmithAPI) (*domain.AppsmithResponse, error) {
	endpoint := fmt.Sprintf("%s/git/discard/app/%s", config.ENDPOINT, config.ApplicationID)
	headers := make(clients.Headers)
	headers["Authorization"] = fmt.Sprintf("Basic %s", config.Authorization)
	headers["branchName"] = branchName
	headers["accept"] = "application/json, text/plain, */*"
	headers["x-requested-by"] = "Appsmith"
	response, err := clients.HandleHTTPRequest[domain.AppsmithResponse](a.HTTPClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodPut,
		URL:     endpoint,
		Body:    nil,
		Headers: &headers,
	})

	return response, err
}
