package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/cerebry"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/go-resty/resty/v2"
)

type CerebryRepo struct {
	Config cerebry.Config
	Client *resty.Client
}

func NewCerebryRepo(config cerebry.Config) *CerebryRepo {
	restyClient := resty.New()
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"jwt-token":    config.PermanentToken,
	})
	restyClient.SetBaseURL(config.BaseURL)
	restyClient.SetTimeout(20 * time.Second)

	return &CerebryRepo{
		Config: config,
		Client: restyClient,
	}
}

type geUserTokenResp struct {
	Token  string `json:"token"`
	Detail string `json:"detail"`
}

func (c *CerebryRepo) GetUserToken(ctx context.Context, userID string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CerebryRepo.GetUserToken")
	defer span.End()

	endpoint := cerebry.EndpointGenerateUserTokenFmt
	url := fmt.Sprintf(string(endpoint), userID)
	resp, err := c.Client.
		R().
		SetContext(ctx).
		Get(url)
	if err != nil || !resp.IsSuccess() {
		if resp.StatusCode() == 404 {
			return "", errors.NewAppError(errors.ErrAPIRespondNotFound, "CerebryRepo.GetUserToken: User is not registered with Cerebry", err)
		}
		return "", errors.New("CerebryRepo.GetUserToken: API error", err)
	}

	respBody := geUserTokenResp{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return "", errors.NewConversionError("CerebryRepo.GetUserToken: Can not unmarshal JSON response", err)
	}

	if respBody.Token == "" {
		return "", errors.New("CerebryRepo.GetUserToken: API does not return a token", nil)
	}

	return respBody.Token, nil
}
