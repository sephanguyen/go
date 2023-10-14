package agora

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/configs"

	"go.uber.org/zap"
)

type agoraClientImpl struct {
	configs.AgoraConfig
	httpClient *http.Client

	logger *zap.Logger
}

func NewAgoraClient(config configs.AgoraConfig, logger *zap.Logger) (chatvendor.ChatVendorClient, error) {
	if config.AppID == "" {
		return nil, fmt.Errorf("[agora]: missing app_id config")
	}

	if config.PrimaryCertificate == "" {
		return nil, fmt.Errorf("[agora]: missing app_certificate config")
	}

	if config.OrgName == "" {
		return nil, fmt.Errorf("[agora]: missing org_name config")
	}

	if config.AppName == "" {
		return nil, fmt.Errorf("[agora]: missing app_name config")
	}

	if config.RestAPI == "" {
		return nil, fmt.Errorf("[agora]: missing REST API address config")
	}

	agoraClient := &agoraClientImpl{
		AgoraConfig: config,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				MaxIdleConnsPerHost: 5,
			},
		},
		logger: logger,
	}

	orgName, appName, err := agoraClient.getAgoraAppInfo()
	if err != nil {
		return nil, fmt.Errorf("[agora] cannot connect to Agora server: [%v]", err)
	}
	logger.Sugar().Infof("[agora]: connected to app [%s#%s]", orgName, appName)

	return agoraClient, nil
}
