package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom"
	domain_zoom "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	cmap "github.com/orcaman/concurrent-map/v2"
	"golang.org/x/sync/semaphore"
)

var semGetConfig = semaphore.NewWeighted(int64(1))

type ExternalConfigService struct {
	secretKey             string
	mapExternalZoomConfig cmap.ConcurrentMap[string, *domain_zoom.ZoomConfigCache]

	configurationClient clients.ConfigurationClientInterface
}

type ExternalConfigServiceInterface interface {
	GetConfigByResource(ctx context.Context) (*domain_zoom.ZoomConfig, error)
}

func InitExternalConfigService(configurationClient clients.ConfigurationClientInterface, secretKey string) *ExternalConfigService {
	return &ExternalConfigService{
		secretKey:             secretKey,
		mapExternalZoomConfig: cmap.New[*domain_zoom.ZoomConfigCache](),
		configurationClient:   configurationClient,
	}
}

func (s *ExternalConfigService) GetConfigByResource(ctx context.Context) (*domain_zoom.ZoomConfig, error) {
	if err := semGetConfig.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer semGetConfig.Release(1)
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	cacheZoomConfig, ok := s.mapExternalZoomConfig.Get(resourcePath)
	now := time.Now()
	if ok {
		tokenExpireTime := cacheZoomConfig.ExpireIn
		if now.Before(*tokenExpireTime) {
			return cacheZoomConfig.ZoomConfig, nil
		}
	}
	newConfig, err := s.loadConfig(ctx, resourcePath)
	if err != nil {
		return nil, err
	}
	// should expire after two hour
	duration := 2 * time.Hour
	expireIn := now.Add(duration)

	s.mapExternalZoomConfig.Set(resourcePath, &domain_zoom.ZoomConfigCache{
		ZoomConfig: newConfig,
		ExpireIn:   &expireIn,
	})
	return newConfig, nil
}

func (s *ExternalConfigService) loadConfig(ctx context.Context, resourcePath string) (*domain_zoom.ZoomConfig, error) {
	getConfigReq := &mpb.GetConfigurationsRequest{
		Keyword: zoom.KeyZoomConfig,
		Paging: &cpb.Paging{
			Limit: 1,
		},
	}
	resp, err := s.configurationClient.GetConfigurations(ctx, getConfigReq)
	if err != nil {
		return nil, err
	}
	items := resp.GetItems()
	if len(resp.GetItems()) < 1 {
		return nil, fmt.Errorf("not found config for org: %s", resourcePath)
	}
	firstItem := items[0]
	zoomConfigEncrypted, err := domain_zoom.InitZoomConfig(firstItem.ConfigValue)
	if err != nil {
		return nil, err
	}
	zoomConfig, err := zoomConfigEncrypted.ToDecrypt(s.secretKey)
	if err != nil {
		return nil, err
	}
	return zoomConfig, nil
}
