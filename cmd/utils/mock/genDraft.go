package mock

import (
	"github.com/manabie-com/backend/internal/draft/repositories"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genDraftRepo(cmd *cobra.Command, args []string) error {
	tools.AddImport("github.com/google/go-github/v41/github")
	structs := map[string][]interface{}{
		"internal/draft/repositories": {
			&repositories.DraftRepo{},
			&repositories.GithubEvent{},
			&repositories.GithubMergeStatusRepo{},
			&repositories.GithubClient{},
		},
	}
	return tools.GenMockStructs(structs)
}

func newGenDraftCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "draft",
		Short: "generate draft repository type",
		Args:  cobra.NoArgs,
		RunE:  genDraftRepo,
	}
}
