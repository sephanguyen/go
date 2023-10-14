package mastermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom"
	zoom_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/configurations"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

var (
	resourcePath string
	isUseZoom    string
	accountID    string
	clientID     string
	clientSecret string
)

func init() {
	bootstrap.RegisterJob("create_config_key_of_zoom_for_partner", createZoomConfigKeyForOrg).
		Desc("create config key of zoom for partner").
		StringVar(&resourcePath, "resourcePath", "", "orgId of partner").
		StringVar(&isUseZoom, "isUseZoom", "", "bool check is enable zoom").
		StringVar(&accountID, "accountID", "", "accountID of zoom provide").
		StringVar(&clientID, "clientID", "", "clientID of zoom provide").
		StringVar(&clientSecret, "clientSecret", "", "clientSecret of zoom provide")
}

func createZoomConfigKeyForOrg(ctx context.Context, cfg configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zLogger := zapLogger.Sugar()
	zLogger.Infof("create zoom configKey for %s", resourcePath)

	// database
	masterDBTrace := rsc.DBWith("mastermgmt")

	createExternalConfigurationHandler := &commands.CreateExternalConfigurationHandler{
		DB:         masterDBTrace,
		ConfigRepo: &repo.ExternalConfigRepo{},
	}
	zoomConfig := &zoom_domain.ZoomConfig{
		AccountID:    accountID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	zoomConfigEncrypted, err := zoomConfig.ToEncrypted(cfg.Zoom.SecretKey)
	if err != nil {
		return fmt.Errorf("encrypt zoom key failed: %w", err)
	}
	zoomConfigEncryptedJSON, err := zoomConfigEncrypted.ToJSONString()
	if err != nil {
		return fmt.Errorf("encrypt zoom key to JSON string failed: %w", err)
	}
	configs := []*mpb.CreateMultiConfigurationsRequest_ExternalConfiguration{
		{Key: zoom.KeyZoomIsEnabled, Value: isUseZoom, ValueType: "boolean"},
		{Key: zoom.KeyZoomConfig, Value: zoomConfigEncryptedJSON, ValueType: "json"},
	}
	newCtx := auth.InjectFakeJwtToken(ctx, resourcePath)
	now := time.Now()
	payload := sliceutils.Map(configs, func(c *mpb.CreateMultiConfigurationsRequest_ExternalConfiguration) *domain.ExternalConfiguration {
		data := &domain.ExternalConfiguration{}

		data.ConfigKey = c.Key
		data.ConfigValue = c.Value
		data.ConfigValueType = c.ValueType
		data.ID = idutil.ULIDNow()
		data.CreatedAt = now
		data.UpdatedAt = now
		return data
	})
	err = createExternalConfigurationHandler.CreateMultiConfigurations(newCtx, payload)
	if err != nil {
		return fmt.Errorf("insert config key failed: %w", err)
	}
	return nil
}
