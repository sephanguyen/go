package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure"
)

type PortForwardClassDoServiceInterface struct{}

type PortForwardClassDoService struct {
	cfg                *configs.ClassDoConfig
	db                 database.Ext
	httpClient         clients.HTTPClientInterface
	ClassDoAccountRepo infrastructure.ClassDoAccountRepo
}

func NewPortForwardClassDoService(cfg *configs.ClassDoConfig, lessonmgmtDB database.Ext, httpClient clients.HTTPClientInterface, classDoAccountRepo infrastructure.ClassDoAccountRepo) *PortForwardClassDoService {
	return &PortForwardClassDoService{
		cfg:                cfg,
		db:                 lessonmgmtDB,
		httpClient:         httpClient,
		ClassDoAccountRepo: classDoAccountRepo,
	}
}

func (pfcs *PortForwardClassDoService) getGeneratedAPIFromID(ctx context.Context, classDoID string) (string, error) {
	classDoAccount, err := pfcs.ClassDoAccountRepo.GetClassDoAccountByID(ctx, pfcs.db, classDoID)
	if err != nil {
		return "", err
	}
	return classDoAccount.ToClassDoAccountDomain(pfcs.cfg.SecretKey).ClassDoAPIKey, nil
}

func (pfcs *PortForwardClassDoService) PortForwardClassDo(ctx context.Context, req *domain.PortForwardClassDoRequest) (*domain.PortForwardClassDoResponse, error) {
	generatedAPI, err := pfcs.getGeneratedAPIFromID(ctx, req.ClassDoID)
	if err != nil {
		return nil, err
	}
	headers := make(clients.Headers)
	headers["Content-Type"] = "application/json"
	headers["x-api-key"] = generatedAPI
	response, err := clients.HandleHTTPRequest[interface{}](pfcs.httpClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodPost,
		URL:     pfcs.cfg.Endpoint,
		Body:    bytes.NewBuffer([]byte(req.Body)),
		Headers: &headers,
	})
	if err != nil {
		return nil, fmt.Errorf("port forward grapql classdo failed: %w", err)
	}

	responseString, err := json.Marshal(*response)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %w", err)
	}

	return &domain.PortForwardClassDoResponse{
		Response: string(responseString),
	}, nil
}
