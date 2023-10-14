package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigUsermgmt() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamUser,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"User.*"},
		},
		{
			Name:      constants.StreamUserDeviceToken,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"UserDeviceToken.*"},
		},
		{
			Name:      constants.StreamOrganization,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"Organization.*"},
		},
		{
			Name:      constants.StreamImportStudentEvent,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"ImportStudent.*"},
		},
		{
			Name:      constants.StreamStaff,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"Staff.*"},
		},
		{
			Name:      constants.StreamSyncUserRegistration,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"SyncUserRegistration.*"},
		},
		{
			Name:      constants.StreamOrderEventLog,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"OrderEventLog.*"},
		},
		{
			Name:      constants.StreamUserGroup,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"UserGroup.*"},
		},
	}

	return arrStreamConfig
}
