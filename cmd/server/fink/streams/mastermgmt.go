package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigMastermgmt() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
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
		{
			Name:      constants.StreamMasterMgmtClass,
			Replicas:  3,
			Subjects:  []string{constants.SubjectMasterMgmtClass},
			Retention: nats_org.InterestPolicy,
		},
		{
			Name:      constants.StreamMasterMgmtReserveClass,
			Replicas:  3,
			Subjects:  []string{constants.SubjectMasterMgmtReserveClass},
			Retention: nats_org.InterestPolicy,
		},
	}

	return arrStreamConfig
}
