package mock

import (
	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/zeus/repositories"

	"github.com/spf13/cobra"
)

func genZeusRepo(cmd *cobra.Command, args []string) error {
	structs := map[string][]interface{}{
		"internal/zeus/repositories": {
			&repositories.ActivityLogRepo{},
		},
	}
	return tools.GenMockStructs(structs)
}

func newGenZeusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zeus",
		Short: "generate zeus repository type",
		Args:  cobra.NoArgs,
		RunE:  genZeusRepo,
	}
}
