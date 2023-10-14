package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigBob() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamCloudConvertJobEvent,
			Replicas:  3,
			Subjects:  []string{"CloudConvertJobEvent.*"},
			Retention: nats_org.InterestPolicy,
		},
		{
			Name:      constants.StreamStudentEventLog,
			Replicas:  3,
			Subjects:  []string{"StudentEventLogs.*"},
			Retention: nats_org.InterestPolicy,
		},
		{
			Name:      constants.StreamSyncLocationTypeUpserted,
			Replicas:  3,
			Subjects:  []string{"SyncLocationType.*"},
			Retention: nats_org.InterestPolicy,
		},
		{
			Name:      constants.StreamSyncLocationUpserted,
			Replicas:  3,
			Subjects:  []string{"SyncLocation.*"},
			Retention: nats_org.InterestPolicy,
		},
	}

	return arrStreamConfig
}
