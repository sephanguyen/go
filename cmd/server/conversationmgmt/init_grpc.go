package conversationmgmt

import (
	agora_usermgmt_grpc "github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/controller/grpc"
	convo_grpc "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/controller/grpc"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"google.golang.org/grpc"
)

func initAgoraUserMgmtServer(
	server grpc.ServiceRegistrar,
	agoraUserMgmtGRPC *agora_usermgmt_grpc.AgoraUserMgmtService,
) {
	cpb.RegisterAgoraUserMgmtServiceServer(server, agoraUserMgmtGRPC)
}

func initConversationModifierServer(
	server grpc.ServiceRegistrar,
	conversationModifierGRPC *convo_grpc.ConversationModifierGRPC,
) {
	cpb.RegisterConversationModifierServiceServer(server, conversationModifierGRPC)
}

func initConversationReaderServer(
	server grpc.ServiceRegistrar,
	conversationReaderGRPC *convo_grpc.ConversationReaderGRPC,
) {
	cpb.RegisterConversationReaderServiceServer(server, conversationReaderGRPC)
}
