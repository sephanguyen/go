package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigPayment() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamSyncGradeEvent,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"SyncGrade.upsert"},
		},
		{
			Name:      constants.StreamSyncLocationUpserted,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectSyncLocationUpserted},
		},
		{
			Name:      constants.StreamOrderEventLog,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectOrderEventLogCreated},
		},
		{
			Name:      constants.StreamStudentCourseEventSync,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectStudentCourseEventSync},
		},
		{
			Name:      constants.StreamOrderWithProductInfoLog,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{constants.SubjectOrderWithProductInfoLogCreated},
		},
	}

	return arrStreamConfig
}
