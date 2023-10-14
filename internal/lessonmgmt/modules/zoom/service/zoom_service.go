package service

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	async "github.com/manabie-com/backend/internal/golibs/asyncawait"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/retry"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	cmap "github.com/orcaman/concurrent-map/v2"
	"golang.org/x/sync/semaphore"
)

var semGetToken = semaphore.NewWeighted(int64(1))
var semGenerateMeetingLink = semaphore.NewWeighted(int64(5))

type ZoomToken struct {
	AccessToken string
	ExpireIn    *time.Time
}

type ZoomGenerateTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type ZoomService struct {
	cfg                   *configs.ZoomConfig
	externalConfigService ExternalConfigServiceInterface
	mapZoomToken          cmap.ConcurrentMap[string, *ZoomToken]
	httpClient            clients.HTTPClientInterface
}

type ZoomServiceInterface interface {
	RetryGenerateZoomLink(ctx context.Context, accountOwner string, req *domain.ZoomGenerateMeetingRequest) (*domain.GenerateZoomLinkResponse, error)
	GenerateMultiZoomLink(ctx context.Context, accountOwner string, req []*domain.ZoomGenerateMeetingRequest) ([]*domain.GenerateZoomLinkResponse, error)
	RetryGetListUsers(ctx context.Context, req *domain.ZoomGetListUserRequest) (*domain.UserZoomResponse, error)
	RetryDeleteZoomLink(ctx context.Context, zoomID string) (bool, error)
}

func InitZoomService(cfg *configs.ZoomConfig,
	externalConfigService ExternalConfigServiceInterface,
	httpClient clients.HTTPClientInterface) *ZoomService {
	return &ZoomService{
		cfg:                   cfg,
		externalConfigService: externalConfigService,
		httpClient:            httpClient,
		mapZoomToken:          cmap.New[*ZoomToken](),
	}
}

func (s *ZoomService) getToken(ctx context.Context) (string, error) {
	if err := semGetToken.Acquire(ctx, 1); err != nil {
		return "", err
	}
	defer semGetToken.Release(1)
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	token, ok := s.mapZoomToken.Get(resourcePath)
	if ok {
		tokenExpireTime := token.ExpireIn
		now := time.Now()
		if now.Before(*tokenExpireTime) {
			return token.AccessToken, nil
		}
	}

	newToken, err := s.generateToken(ctx)
	s.mapZoomToken.Set(resourcePath, newToken)

	if err != nil {
		return "", fmt.Errorf("getToken fail: %w", err)
	}
	return newToken.AccessToken, nil
}

func (s *ZoomService) generateToken(ctx context.Context) (*ZoomToken, error) {
	cfgZoom, err := s.externalConfigService.GetConfigByResource(ctx)
	if err != nil {
		return nil, err
	}
	requestGenerateToken := fmt.Sprintf("%s?grant_type=account_credentials&account_id=%s", s.cfg.EndpointAuth, cfgZoom.AccountID)
	pairKey := fmt.Sprintf("%s:%s", cfgZoom.ClientID, cfgZoom.ClientSecret)
	base64PairKey := b64.StdEncoding.EncodeToString([]byte(pairKey))
	basicToken := fmt.Sprintf("Basic %s", base64PairKey)
	headers := make(clients.Headers)
	headers["Authorization"] = basicToken

	response, err := clients.HandleHTTPRequest[ZoomGenerateTokenResponse](s.httpClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodPost,
		URL:     requestGenerateToken,
		Body:    nil,
		Headers: &headers,
	})
	if err != nil {
		return nil, fmt.Errorf("generateToken fail: %w", err)
	}
	// expire after 45minus
	now := time.Now()
	duration := 45 * time.Minute
	expireIn := now.Add(duration)
	zoomToken := &ZoomToken{AccessToken: response.AccessToken, ExpireIn: &expireIn}
	return zoomToken, nil
}

func (s *ZoomService) generateZoomLink(ctx context.Context, accountOwner string, req *domain.ZoomGenerateMeetingRequest) (*domain.GenerateZoomLinkResponse, error) {
	if err := semGenerateMeetingLink.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer semGenerateMeetingLink.Release(1)
	token, err := s.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("GenerateZoomLink fail: %w", err)
	}
	requestGenerateZoomLink := fmt.Sprintf("%s/users/%s/meetings", s.cfg.Endpoint, accountOwner)
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	authorizationToken := fmt.Sprintf("Bearer %s", token)

	headers := make(clients.Headers)
	headers["Authorization"] = authorizationToken
	headers["User-Agent"] = "Zoom-api-Jwt-Request"
	headers["content-type"] = "application/json"

	data, err := clients.HandleHTTPRequest[domain.ZoomGenerateMeetingResponse](s.httpClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodPost,
		URL:     requestGenerateZoomLink,
		Body:    bytes.NewBuffer(jsonReq),
		Headers: &headers,
	})

	if err != nil {
		return nil, retry.NewStop(fmt.Errorf("GenerateZoomLink fail: %w", err))
	}
	// Invalid access token. if invalid need to refresh token
	if data.Code == 124 {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		s.mapZoomToken.Remove(resourcePath)
	}
	if data.Code != 0 {
		return nil, fmt.Errorf("GenerateZoomLink fail: %s", data.Message)
	}

	return &domain.GenerateZoomLinkResponse{
		ZoomID:      data.ID,
		URL:         data.URL,
		Occurrences: data.Occurrences,
	}, nil
}

func (s *ZoomService) getListUsers(ctx context.Context, req *domain.ZoomGetListUserRequest) (*domain.UserZoomResponse, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getListUsers fail: %w", err)
	}
	requestGetListUser := fmt.Sprintf("%s/users?page_size=%d&page_number=%d", s.cfg.Endpoint, req.PageSize, req.PageNumber)
	if err != nil {
		return nil, err
	}

	authorizationToken := fmt.Sprintf("Bearer %s", token)

	headers := make(clients.Headers)
	headers["Authorization"] = authorizationToken
	headers["User-Agent"] = "Zoom-api-Jwt-Request"
	headers["content-type"] = "application/json"

	data, err := clients.HandleHTTPRequest[domain.UserZoomResponse](s.httpClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodGet,
		URL:     requestGetListUser,
		Headers: &headers,
	})

	if err != nil {
		return nil, retry.NewStop(fmt.Errorf("getListUsers fail: %w", err))
	}
	// Invalid access token. if invalid need to refresh token
	if data.Code == 124 {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		s.mapZoomToken.Remove(resourcePath)
	}
	if data.Code != 0 {
		return nil, fmt.Errorf("getListUsers fail: %s", data.Message)
	}

	return data, nil
}

func (s *ZoomService) deleteZoomLink(ctx context.Context, zoomID string) (bool, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		return false, fmt.Errorf("deleteZoomLink fail: %w", err)
	}
	requestDeleteZoomLink := fmt.Sprintf("%s/meetings/%s", s.cfg.Endpoint, zoomID)
	authorizationToken := fmt.Sprintf("Bearer %s", token)

	headers := make(clients.Headers)
	headers["Authorization"] = authorizationToken
	headers["User-Agent"] = "Zoom-api-Jwt-Request"
	headers["content-type"] = "application/json"

	data, err := clients.HandleHTTPRequest[domain.DeleteZoomResponse](s.httpClient, &clients.RequestInput{
		Ctx:     ctx,
		Method:  http.MethodDelete,
		URL:     requestDeleteZoomLink,
		Headers: &headers,
	})
	if err != nil {
		return false, retry.NewStop(fmt.Errorf("deleteZoomLink fail: %w", err))
	}
	if data != nil {
		// Invalid access token. if invalid need to refresh token
		if data.Code == 124 {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			s.mapZoomToken.Remove(resourcePath)
		}
		if data.Code == 3001 {
			// should by pass when the Meeting is not found or has expired.
			return true, nil
		}
		if data.Code != 0 {
			return false, fmt.Errorf("deleteZoomLink fail: %s", data.Message)
		}
	}

	return true, nil
}

func (s *ZoomService) RetryGenerateZoomLink(ctx context.Context, accountOwner string, req *domain.ZoomGenerateMeetingRequest) (*domain.GenerateZoomLinkResponse, error) {
	return retry.Retry(3, time.Microsecond, func() (*domain.GenerateZoomLinkResponse, error) {
		return s.generateZoomLink(ctx, accountOwner, req)
	})
}

func (s *ZoomService) RetryGetListUsers(ctx context.Context, req *domain.ZoomGetListUserRequest) (*domain.UserZoomResponse, error) {
	return retry.Retry(3, time.Microsecond, func() (*domain.UserZoomResponse, error) {
		return s.getListUsers(ctx, req)
	})
}

func (s *ZoomService) RetryDeleteZoomLink(ctx context.Context, zoomID string) (bool, error) {
	return retry.Retry(3, time.Microsecond, func() (bool, error) {
		return s.deleteZoomLink(ctx, zoomID)
	})
}

func (s *ZoomService) GenerateMultiZoomLink(ctx context.Context, accountOwner string, req []*domain.ZoomGenerateMeetingRequest) ([]*domain.GenerateZoomLinkResponse, error) {
	var wg sync.WaitGroup
	totalLinkShouldBeGen := len(req)
	wg.Add(totalLinkShouldBeGen)
	futures := sliceutils.Map(req, func(request *domain.ZoomGenerateMeetingRequest) async.Future {
		return async.Exec(func() (interface{}, error) {
			defer wg.Done()
			return s.RetryGenerateZoomLink(ctx, accountOwner, request)
		})
	})

	wg.Wait()
	zoomLinks := make([]*domain.GenerateZoomLinkResponse, 0, totalLinkShouldBeGen)
	for i := 0; i < totalLinkShouldBeGen; i++ {
		f := futures[i]
		result, err := f.Await()
		if err != nil {
			return nil, fmt.Errorf("GenerateMultiZoomLink fail: %s", err)
		}
		zoomLinks = append(zoomLinks, result.(*domain.GenerateZoomLinkResponse))
	}

	return zoomLinks, nil
}
