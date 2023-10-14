package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/enigma/repositories"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genEnigmaRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"partner_sync_data_log":       &repositories.PartnerSyncDataLogRepo{},
		"partner_sync_data_log_split": &repositories.PartnerSyncDataLogSplitRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "enigma", repos)

	interfaces := map[string][]string{
		"internal/enigma/controllers": {
			"InternalClient",
			"UserServiceClient",
		},
	}

	return tools.GenMockInterfaces(interfaces)
}

func newGenEnigmaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enigma",
		Short: "Generate mocks for enigma",
		Args:  cobra.ExactArgs(1),
		RunE:  genEnigmaRepo,
	}
}
