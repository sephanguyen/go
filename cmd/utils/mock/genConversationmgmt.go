package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genConversationmgmtMock(_ *cobra.Command, args []string) error {
	repos := map[string]interface{}{}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "conversationmgmt", repos)

	agoraUsermgmtRepos := map[string]interface{}{
		"agora_user":      &repositories.AgoraUserRepo{},
		"user_basic_info": &repositories.UserBasicInfoRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "modules/agora_usermgmt/infrastructure/repositories"), "conversationmgmt/modules/agora_usermgmt", agoraUsermgmtRepos)

	conversationMgmtRepos := map[string]interface{}{
		"agora_user":          &postgres.AgoraUserRepo{},
		"conversation":        &postgres.ConversationRepo{},
		"conversation_member": &postgres.ConversationMemberRepo{},
		"internal_admin_user": &postgres.InternalAdminUserRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "modules/conversation/infrastructure/postgres"), "conversationmgmt/modules/conversation", conversationMgmtRepos)

	structs := map[string][]interface{}{}
	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/conversationmgmt/modules/conversation/core/port/service": {
			"ConversationModifierService",
			"ConversationReaderService",
			"NotificationHandler",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenConversationmgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "conversationmgmt [../../mock/conversationmgmt]",
		Short: "Generate mock repositories for conversationmgmt",
		Args:  cobra.ExactArgs(1),
		RunE:  genConversationmgmtMock,
	}
}
