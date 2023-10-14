package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genEntryexitmgmtRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"student_qr":                &repositories.StudentQRRepo{},
		"student_entryexit_records": &repositories.StudentEntryExitRecordsRepo{},
		"entryexit_queue":           &repositories.EntryExitQueueRepo{},
		"student_parent":            &repositories.StudentParentRepo{},
		"student":                   &repositories.StudentRepo{},
		"user":                      &repositories.UserRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "entryexitmgmt", repos)

	interfaces := map[string][]string{
		"internal/entryexitmgmt/services/filestore": {
			"FileStore",
		},
		"internal/entryexitmgmt/services/uploader": {
			"Uploader",
		},
	}
	if err := tools.GenMockInterfaces(interfaces); err != nil {
		return err
	}

	return nil
}

func newGenEntryexitmgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "entryexitmgmt [../../mock/entryexitmgmt]",
		Short: "generate entryexitmgmt repository type",
		Args:  cobra.MaximumNArgs(1),
		RunE:  genEntryexitmgmtRepo,
	}
}
