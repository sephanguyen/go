package spike

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	"github.com/manabie-com/backend/internal/spike/configurations"

	"go.uber.org/zap"
)

// TODO: replace the returned interface with a more generic provider interface
func initEmailProvider(c configurations.Config, log *zap.Logger) (sgClient sendgrid.SendGridClient) {
	var err error

	if c.Common.Environment != localEnv {
		sgClient, err = sendgrid.NewSendGridClient(c.SendGrid)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed init SendGridClient: %+v", err))
		}

		return
	}

	sgClient = sendgrid.NewSendGridMock()

	return
}
