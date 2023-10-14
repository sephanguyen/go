package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigYasuo() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamSyncUserCourse,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"SyncUserCourse.*"},
		},
		// {
		// 	Name:      constants.StreamSyncUserRegistration,
		// 	Retention: nats_org.InterestPolicy,
		// 	Replicas:  3,
		// 	Subjects:  []string{"SyncUserRegistration.*"},
		// },
		{
			Name:      constants.StreamSyncMasterRegistration,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"SyncMasterRegistration.*"},
		},
	}

	return arrStreamConfig
}
