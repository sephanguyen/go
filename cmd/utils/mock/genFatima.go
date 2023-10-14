package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genFatimaRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"package":                     &repositories.PackageRepo{},
		"student_package":             &repositories.StudentPackageRepo{},
		"student_package_access_path": &repositories.StudentPackageAccessPathRepo{},
		"student_package_class":       &repositories.StudentPackageClassRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "fatima", repos)

	return nil
}

func newGenFatimaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fatima [../../mock/fatima]",
		Short: "generate fatima repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genFatimaRepo,
	}
}
