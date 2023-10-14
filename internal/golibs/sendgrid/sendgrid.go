package sendgrid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	SGHost              = "https://api.sendgrid.com"
	SGSendEmailEndpoint = "/v3/mail/send"
)

// nolint:revive
type SendGridClient interface {
	Send(email *mail.SGMailV3) (string, error)
	SendWithContext(ctx context.Context, email *mail.SGMailV3) (string, error)
	AuthenticateHTTPRequest(header http.Header, payload []byte) (bool, error)
}

type sgClientImpl struct {
	client sendgrid.Client
	configs.SendGridConfig
}

func NewSendGridClient(securityCfg configs.SendGridConfig) (SendGridClient, error) {
	if securityCfg.APIKey == "" || securityCfg.PublicKey == "" {
		return nil, fmt.Errorf("missing SendGrid security configurations")
	}

	return &sgClientImpl{
		SendGridConfig: securityCfg,
		client:         *sendgrid.NewSendClient(securityCfg.APIKey),
	}, nil
}

func (sgc *sgClientImpl) Send(email *mail.SGMailV3) (string, error) {
	response, err := sgc.client.Send(email)
	if err != nil {
		return "", fmt.Errorf("failed SendGrid Send: %v", err)
	}

	return HandleResponse(response)
}

func (sgc *sgClientImpl) SendWithContext(ctx context.Context, email *mail.SGMailV3) (string, error) {
	response, err := sgc.client.SendWithContext(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed SendGrid SendWithContext: %v", err)
	}

	return HandleResponse(response)
}
