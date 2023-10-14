package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/commands"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure/repositories"

	"github.com/spf13/cobra"
)

func genSpikeRepo(_ *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"email":                 &repositories.EmailRepo{},
		"email_recipient":       &repositories.EmailRecipientRepo{},
		"email_recipient_event": &repositories.EmailRecipientEventRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "modules/email/infrastructure/repositories"), "spike/modules/email", repos)

	structs := map[string][]interface{}{
		"internal/spike/modules/email/application/commands": {
			&commands.CreateEmailHandler{},
			&commands.UpsertEmailEventHandler{},
		},
	}

	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/spike/modules/email/metrics": {
			"EmailMetrics",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenSpikeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spike [../../mock/spike]",
		Short: "Generate mock repositories for spike",
		Args:  cobra.ExactArgs(1),
		RunE:  genSpikeRepo,
	}
}
