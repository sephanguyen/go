package constants

import pb "github.com/manabie-com/backend/pkg/genproto/tom"

var SystemMessageOnly = [4]pb.CodesMessageType{
	pb.CODES_MESSAGE_TYPE_ALLOW_STUDENT_TO_SPEAK,
	pb.CODES_MESSAGE_TYPE_PROHIBIT_STUDENT_TO_SPEAK,
	pb.CODES_MESSAGE_TYPE_ALLOW_STUDENT_TO_CHAT,
	pb.CODES_MESSAGE_TYPE_PROHIBIT_STUDENT_TO_CHAT,
}

const (
	UserDeviceTokenNats            = "user_device_token"
	UserDeviceTokenNatsQueueGroup  = "user_device_token_queue"
	UserDeviceTokenNatsDurableName = "user_device_token_durable"

	SubjectCreateStudentQuestionNats             = "student_questions"
	QueueGroupSubjectCreateStudentQuestionNats   = "queue_student_questions"
	DurableGroupSubjectCreateStudentQuestionNats = "durable_student_questions"

	SubjectClassEventNats             = "class_event"
	QueueGroupSubjectClassEventNats   = "queue_class_event"
	DurableGroupSubjectClassEventNats = "durable_class_event"

	MessagingFileNotificationContent = "ファイルを受信しました" //TODO: align with mobile to use localization

	FcmKeyNotificationType = "notification_type"
	FcmKeyItemID           = "item_id"
	FcmKeyClickAction      = "click_action"
	FcmKeyConversationName = "conversation_name"
	FcmKeyMessageContent   = "message_content"
)

const (
	ConversationStatusActive   = "CONVERSATION_STATUS_ACTIVE"
	ConversationStatusInactive = "CONVERSATION_STATUS_INACTIVE"

	ChatConfigKeyStudent = "communication.chat.enable_student_chat"
	ChatConfigKeyParent  = "communication.chat.enable_parent_chat"

	ChatConfigKeyStudentV2 = "communication.chat.enable_student_chat_v2"
	ChatConfigKeyParentV2  = "communication.chat.enable_parent_chat_v2"

	LocalEnv   = "local"
	StagingEnv = "stag"
	UATEnv     = "uat"
	ProdEnv    = "prod"

	ChatThreadLaunchingFeatureFlag = "Communication_Chat_ChatThreadLaunching_Phase2"
)
