package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigFatima() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamSyncJprepStudentPackageEventNats,
			Retention: nats_org.InterestPolicy,
			Subjects:  []string{constants.SubjectSyncJprepStudentPackageEventNats},
			Replicas:  3,
		},
	}

	return arrStreamConfig
}
