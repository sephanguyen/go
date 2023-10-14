package sendgrid

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/sendgrid/sendgrid-go"
	"github.com/stretchr/testify/assert"
)

func Test_InitSendGrid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name         string
		SecurityCfg  configs.SendGridConfig
		ExpectErr    error
		ExpectClient SendGridClient
	}{
		{
			Name: "happy case",
			SecurityCfg: configs.SendGridConfig{
				APIKey:    "API_KEY",
				PublicKey: "PUBLIC_KEY",
			},
			ExpectErr: nil,
			ExpectClient: &sgClientImpl{
				SendGridConfig: configs.SendGridConfig{
					APIKey:    "API_KEY",
					PublicKey: "PUBLIC_KEY",
				},
				client: *sendgrid.NewSendClient("API_KEY"),
			},
		},
		{
			Name: "missing config",
			SecurityCfg: configs.SendGridConfig{
				APIKey:    "API_KEY",
				PublicKey: "",
			},
			ExpectErr:    fmt.Errorf("missing SendGrid security configurations"),
			ExpectClient: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			client, err := NewSendGridClient(tc.SecurityCfg)
			assert.Equal(t, tc.ExpectErr, err)
			assert.Equal(t, tc.ExpectClient, client)
		})
	}
}
