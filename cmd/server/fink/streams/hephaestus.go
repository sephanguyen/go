package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigHephaestus() []*nats_org.StreamConfig {
	var arrStreamConfig = []*nats_org.StreamConfig{
		{
			Name:      constants.StreamDebeziumIncrementalSnapshot,
			Retention: nats_org.InterestPolicy,
			Replicas:  3,
			Subjects:  []string{"DebeziumIncrementalSnapshot.*"},
		},
	}

	return arrStreamConfig
}
