package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigVirtualClassroom() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamUpcomingLiveLessonNotification,
			Replicas:  3,
			Subjects:  []string{constants.SubjectUpcomingLiveLessonNotification},
			Retention: nats_org.InterestPolicy,
		},
		{
			Name:      constants.StreamLiveRoom,
			Replicas:  3,
			Subjects:  []string{constants.SubjectLiveRoomUpdated},
			Retention: nats_org.InterestPolicy,
		},
	}

	return arrStreamConfig
}
