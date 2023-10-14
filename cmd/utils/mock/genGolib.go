package mock

import (
	"github.com/manabie-com/backend/internal/golibs/cloudconvert"
	"github.com/manabie-com/backend/internal/golibs/curl"
	fileio "github.com/manabie-com/backend/internal/golibs/io"
	"github.com/manabie-com/backend/internal/golibs/speeches"
	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"

	"github.com/spf13/cobra"
)

func genGolib(cmd *cobra.Command, args []string) error {
	structs := map[string][]interface{}{
		"internal/golibs/cloudconvert": {
			&cloudconvert.Service{},
		},
		"internal/golibs/curl": {
			&curl.HTTP{},
		},
		"internal/golibs/speeches": {
			&speeches.Text2SpeechBuilder{},
			&speeches.Text2SpeechClient{},
		},
		"internal/golibs/whiteboard": {
			&whiteboard.Service{},
		},
		"internal/golibs/io": {
			&fileio.FileUtils{},
		},
	}
	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/golibs/auth/multitenant": {
			"Tenant",
			"TenantClient",
			"TenantIdentifier",
			"TenantInfo",
			"TenantManager",
		},
		"internal/golibs/auth/user": {
			"User",
		},
		"internal/golibs/kafkaconnect": {
			"ConnectClient",
			"KafkaAdmin",
		},

		"internal/golibs/auth": {
			"BobUserModifierServiceClient",
			"GCPPager",
			"GCPTenantClient",
			"GCPTenantManager",
			"GCPUtils",
			"IdentityService",
			"KeyCloakClient",
			"ScryptHash",
		},
		"internal/golibs/bootstrap": {"Databaser", "Elasticer", "NATSJetstreamer"},
		"internal/golibs/caching": {
			"LocalCacher",
		},
		"internal/golibs/ci/coverage": {
			"SendCoverageServiceClient",
		},
		"internal/golibs/curl": {
			"IHTTP",
		},
		"internal/golibs/database": {
			"BatchResults",
			"Entity",
			"Ext",
			"QueryExecer",
			"Row",
			"Rows",
			"Tx",
		},
		"internal/golibs/elastic": {
			"SearchFactory",
		},
		"internal/golibs/healthcheck": {
			"Pinger",
		},
		"internal/golibs/firebase": {
			"AuthClient",
			"AuthUtils",
			"NotificationPusher",
		},
		"internal/golibs/nats": {
			// "BusFactory", // library error, see https://github.com/vektra/mockery/issues/249
			// "JetStreamManagement", // library error
		},
		"internal/tom/infra/stress": {
			"GrpcClient",
			"ClientStream",
		},
		"internal/golibs/vision": {
			"Factory",
		},
		"internal/golibs/mathpix": {
			"Factory",
		},
		"internal/golibs/learnosity": {
			"Init",
			"DataAPI",
			"HTTP",
		},
		"internal/golibs/alert": {
			"SlackFactory",
		},
		"internal/golibs/kafka": {
			"KafkaManagement",
		},
		"internal/golibs/sendgrid": {
			"SendGridClient",
		},
		"internal/golibs/chatvendor": {
			"ChatVendorClient",
		},
		"internal/golibs/chatvendor/agora": {
			"WebhookVerifier",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenGolibCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "golibs",
		Short: "generate golibs mock structs",
		Args:  cobra.NoArgs,
		RunE:  genGolib,
	}
}
