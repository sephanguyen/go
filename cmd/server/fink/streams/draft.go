package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigDraft() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamCleanDataTestEventNats,
			Replicas:  3,
			Subjects:  []string{constants.SubjectCleanDataTestEventNats},
			Retention: nats_org.InterestPolicy,
		},
	}

	return arrStreamConfig
}
