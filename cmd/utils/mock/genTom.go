package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/tom/app/support"
	"github.com/manabie-com/backend/internal/tom/repositories"

	"github.com/spf13/cobra"
)

func genTomRepo(cmd *cobra.Command, args []string) error {
	tools.AddImport("github.com/manabie-com/backend/internal/tom/domain/core")
	tools.AddImport("github.com/manabie-com/backend/internal/tom/domain/lesson")
	tools.AddImport("github.com/manabie-com/backend/internal/tom/domain/support")
	repos := map[string]interface{}{
		"conversation":                &repositories.ConversationRepo{},
		"conversation_members":        &repositories.ConversationMemberRepo{},
		"conversation_lessons":        &repositories.ConversationLessonRepo{},
		"message":                     &repositories.MessageRepo{},
		"user_device_token":           &repositories.UserDeviceTokenRepo{},
		"online_user":                 &repositories.OnlineUserRepo{},
		"conversation_student":        &repositories.ConversationStudentRepo{},
		"conversation_search":         &repositories.SearchRepo{},
		"conversation_location":       &repositories.ConversationLocationRepo{},
		"location":                    &repositories.LocationRepo{},
		"private_conversation_lesson": &repositories.PrivateConversationLessonRepo{},
		"users":                       &repositories.UsersRepo{},
		"granted_permissions":         &repositories.GrantedPermissionsRepo{},
		"user_group_member":           &repositories.UserGroupMembersRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "tom", repos)

	structs := map[string][]interface{}{
		"internal/tom/app/support": {
			&support.LocationConfigResolver{},
		},
	}

	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/tom/app/core": {
			"ChatInfra",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenTomCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tom [../../mock/tom]",
		Short: "Generate mock repositories for Tom",
		Args:  cobra.ExactArgs(1),
		RunE:  genTomRepo,
	}
}
