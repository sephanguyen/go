package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/usermgmt/modules/auth/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/auth/core/service"

	"github.com/spf13/cobra"
)

func genAuthMock(_ *cobra.Command, args []string) error {
	tools.AddImport("github.com/manabie-com/backend/internal/usermgmt/modules/auth/core/entity")
	repos := map[string]interface{}{
		"organization": &repository.OrganizationRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "auth", repos)

	services := map[string][]interface{}{
		"internal/auth/service": {&service.DomainAuthService{}},
	}
	return tools.GenMockStructs(services)
}

func newGenAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth [../../mock/auth]",
		Short: "generate auth repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genAuthMock,
	}
}
