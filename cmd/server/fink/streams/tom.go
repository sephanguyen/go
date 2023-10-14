package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigTom() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamChatMigration,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"chat_migrate.*"},
		},
		{
			Name:      constants.StreamChatMessage,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"ChatMessage.*"},
		},
		{
			Name:      constants.StreamLessonChat,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"LessonChat.*"},
		},
		{
			Name:      constants.StreamLesson,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"Lesson.*"},
		},
		{
			Name:      constants.StreamSyncStudentLessons,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"SyncStudentLessonsConversations.Synced"},
		},
		{
			Name:      constants.StreamChat,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"chat.chat.>"},
		},
		{
			Name:      constants.StreamESConversation,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectESConversation},
		},
	}

	return arrStreamConfig
}
