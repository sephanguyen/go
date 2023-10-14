package conversationmgmt

import (
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/configurations"
	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora"

	"go.uber.org/zap"
)

func initChatProvider(c configurations.Config, log *zap.Logger) (chatVendor chatvendor.ChatVendorClient) {
	var err error

	if c.Common.Environment != localEnv {
		chatVendor, err = agora.NewAgoraClient(c.Agora, log)
		if err != nil {
			log.Fatal(fmt.Sprintf("cannot init AgoraClient: [%v]", err))
		}
		return
	}
	chatVendor = agora.NewMockAgoraClient()
	return
}
