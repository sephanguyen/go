package agora

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
)

func (a *agoraClientImpl) getAgoraAppInfo() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// Get App Info endpoint: GET /
	endpoint := "/"

	getAppInfoRes := &dto.GetAgoraAppInfoResponse{}
	err := a.doRequest(ctx, MethodGet, endpoint, GetAgoraCommonHeader(), nil, getAppInfoRes)
	if err != nil {
		return "", "", err
	}

	if len(getAppInfoRes.Entities) > 0 {
		appInfo := getAppInfoRes.Entities[0]

		if appInfo.ApplicationName == "" || appInfo.OrganizationName == "" {
			return "", "", fmt.Errorf("empty Agora app")
		}

		return appInfo.OrganizationName, appInfo.ApplicationName, nil
	}

	return "", "", fmt.Errorf("error empty Agora app: [%v]", err)
}
